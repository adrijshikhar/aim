package oauth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

// OpenBrowser attempts to open targetURL in the system's default web browser.
func OpenBrowser(targetURL string) error {
	targetURL = strings.TrimSpace(targetURL)
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		return fmt.Errorf("invalid URL scheme: must be http or https")
	}
	if err := browser.OpenURL(targetURL); err == nil {
		return nil
	}
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", "--", targetURL).Start()
	case "linux":
		return exec.Command("xdg-open", targetURL).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL).Start()
	}
	return fmt.Errorf("failed to open browser on %s", runtime.GOOS)
}

// Authenticate starts an ephemeral loopback HTTP server and executes an OAuth 2.0 PKCE flow.
func Authenticate(ctx context.Context, cfg ProviderConfig, openBrowser bool) (*TokenResult, error) {
	verifier, _, err := GeneratePKCE()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE challenge: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to bind loopback listener: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	state, err := GenerateRandomState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate random state: %w", err)
	}

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", handleCallback(state, codeChan, errChan))

	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(listener)
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	var opts []oauth2.AuthCodeOption
	opts = append(opts, oauth2.S256ChallengeOption(verifier))
	for k, v := range cfg.ExtraAuthParams {
		opts = append(opts, oauth2.SetAuthURLParam(k, v))
	}

	oauthConfig := &oauth2.Config{
		ClientID: cfg.ClientID,
		Endpoint: oauth2.Endpoint{
			AuthURL: cfg.AuthURL,
		},
		RedirectURL: redirectURI,
		Scopes:      cfg.Scopes,
	}
	fullAuthURL := strings.TrimSpace(oauthConfig.AuthCodeURL(state, opts...))
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#61afef")).
		Padding(0, 1).
		Render(fmt.Sprintf("%s\n%s",
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#61afef")).Render("🌐 Opening browser for authorization:"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#56b6c2")).Render(fullAuthURL),
		))
	fmt.Printf("\n%s\n\n", card)
	if openBrowser {
		if err := OpenBrowser(fullAuthURL); err != nil {
			note := lipgloss.NewStyle().Foreground(lipgloss.Color("#e5c07b")).Render(fmt.Sprintf("Note: Could not open browser automatically: %v\nPlease copy and open the URL above.", err))
			fmt.Printf("%s\n\n", note)
		}
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(2 * time.Minute):
		return nil, fmt.Errorf("authentication timed out after 2 minutes")
	case err := <-errChan:
		return nil, err
	case code := <-codeChan:
		return ExchangeCode(ctx, cfg, code, verifier, redirectURI)
	}
}

func handleCallback(state string, codeChan chan<- string, errChan chan<- error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != state {
			http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
			select {
			case errChan <- fmt.Errorf("CSRF state mismatch"):
			default:
			}
			return
		}
		if errParam := q.Get("error"); errParam != "" {
			http.Error(w, "Authorization denied", http.StatusBadRequest)
			select {
			case errChan <- fmt.Errorf("authorization denied: %s", errParam):
			default:
			}
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			select {
			case errChan <- fmt.Errorf("missing authorization code"):
			default:
			}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<html><body style="font-family: sans-serif; text-align: center; padding: 50px;">
			<h2 style="color: #2e7d32;">Authentication Successful!</h2>
			<p>You can close this tab and return to your terminal.</p>
		</body></html>`)
		select {
		case codeChan <- code:
		default:
		}
	}
}
