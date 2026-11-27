package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/merge"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/service"
	"github.com/aim-cli/aim/internal/share"
	"github.com/aim-cli/aim/internal/web"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

const defaultWebPort = 18480

var openBrowserURL = browser.OpenURL

type webPIDInfo struct {
	PID       int       `json:"pid"`
	Port      int       `json:"port"`
	StartedAt time.Time `json:"started_at"`
}

func resolveDefaultPort() int {
	if envPort := os.Getenv("AIM_WEB_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 && p <= 65535 {
			return p
		}
	}
	return defaultWebPort
}

func webPIDPath() string {
	return filepath.Join(config.BaseDir(), "web.pid")
}

func webLogPath() string {
	return filepath.Join(config.BaseDir(), "web.log")
}

func readWebPID() (*webPIDInfo, error) {
	data, err := os.ReadFile(webPIDPath())
	if err != nil {
		return nil, err
	}
	var info webPIDInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func writeWebPID(pid, port int) error {
	info := webPIDInfo{
		PID:       pid,
		Port:      port,
		StartedAt: time.Now(),
	}
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return merge.AtomicWrite(webPIDPath(), data, 0o600)
}

func removeOwnWebPID() error {
	info, err := readWebPID()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info != nil && info.PID != os.Getpid() {
		return nil
	}
	return removeWebPID()
}

func removeWebPID() error {
	err := os.Remove(webPIDPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func checkServerHealth(ctx context.Context, port int) bool {
	if ctx == nil {
		ctx = context.Background()
	}
	reqCtx, cancel := context.WithTimeout(ctx, 400*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/api/health", port), nil)
	if err != nil {
		return false
	}
	client := http.Client{Timeout: 400 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusOK
}

func newWebCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var (
		port       int
		devMode    bool
		noOpen     bool
		foreground bool
	)

	cmd := &cobra.Command{
		Use:   "web [flags]",
		Short: "Start the AIM web dashboard",
		Long: `Start the AIM web dashboard HTTP server with REST APIs and embedded UI.

By default, "aim web" runs as a background process and returns to your shell immediately.
To view or control the server, use:
  aim web status
  aim web stop

Examples:
  aim web
  aim web --no-open
  aim web --port 18480
  aim web --foreground`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if foreground {
				return runWebForeground(cmd, reg, pm, port, devMode, noOpen)
			}
			return startWebBackground(cmd.Context(), cmd.OutOrStdout(), port, devMode, noOpen)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", resolveDefaultPort(), "Port to listen on (default 18480)")
	cmd.Flags().BoolVar(&devMode, "dev", false, "Run in development mode (CORS and proxy to Vite dev server)")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open browser automatically")
	cmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run in foreground attached to terminal")

	cmd.AddCommand(
		newWebStartCmd(reg, pm),
		newWebStopCmd(),
		newWebStatusCmd(),
	)

	return cmd
}

func newWebStartCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var (
		port       int
		devMode    bool
		noOpen     bool
		foreground bool
	)

	cmd := &cobra.Command{
		Use:   "start [flags]",
		Short: "Start the AIM web dashboard server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if foreground {
				return runWebForeground(cmd, reg, pm, port, devMode, noOpen)
			}
			return startWebBackground(cmd.Context(), cmd.OutOrStdout(), port, devMode, noOpen)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", resolveDefaultPort(), "Port to listen on (default 18480)")
	cmd.Flags().BoolVar(&devMode, "dev", false, "Run in development mode")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open browser automatically")
	cmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run in foreground attached to terminal")

	return cmd
}

func newWebStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the background AIM web dashboard server",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			info, err := readWebPID()
			if err != nil {
				if os.IsNotExist(err) {
					_, _ = fmt.Fprintln(out, "AIM Web Dashboard is not running.")
					return nil
				}
				return fmt.Errorf("failed to read dashboard PID file: %w", err)
			}
			if !isProcessAlive(info.PID) {
				_ = removeWebPID()
				_, _ = fmt.Fprintln(out, "AIM Web Dashboard is not running.")
				return nil
			}

			if err := terminateProcess(info.PID); err != nil {
				return fmt.Errorf("failed to terminate dashboard process (PID %d): %w", info.PID, err)
			}
			_ = removeWebPID()
			_, _ = fmt.Fprintf(out, "AIM Web Dashboard stopped (PID %d)\n", info.PID)
			return nil
		},
	}
}

func newWebStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check the running status of the AIM web dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			info, err := readWebPID()
			if err == nil && isProcessAlive(info.PID) && checkServerHealth(cmd.Context(), info.Port) {
				_, _ = fmt.Fprintf(out, "● AIM Web Dashboard is running at http://localhost:%d (PID %d)\n", info.Port, info.PID)
				return nil
			}
			_, _ = fmt.Fprintln(out, "○ AIM Web Dashboard is stopped.")
			return nil
		},
	}
}

func startWebBackground(ctx context.Context, out io.Writer, port int, devMode, noOpen bool) error {
	if out == nil {
		out = os.Stdout
	}
	// 1. Check if already running and responsive
	if info, err := readWebPID(); err == nil {
		if isProcessAlive(info.PID) && checkServerHealth(ctx, info.Port) {
			serverURL := fmt.Sprintf("http://localhost:%d", info.Port)
			_, _ = fmt.Fprintf(out, "AIM Web Dashboard is already running at %s (PID %d)\n", serverURL, info.PID)
			if !noOpen {
				_ = openBrowserURL(serverURL)
			}
			return nil
		}
		// Stale PID file
		_ = removeWebPID()
	}

	// 2. Resolve executable
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}

	args := []string{"web", "--foreground", "--port", strconv.Itoa(port), "--no-open"}
	if devMode {
		args = append(args, "--dev")
	}

	child := exec.Command(exe, args...) //nolint:noctx,gosec // background web daemon re-executes current aim binary and must outlive parent CLI command
	setSysProcAttr(child)

	closeLog := attachWebBackgroundLog(child)
	defer closeLog()

	if err := child.Start(); err != nil {
		return fmt.Errorf("failed to spawn web dashboard process: %w", err)
	}

	childErrChan := make(chan error, 1)
	go func() {
		childErrChan <- child.Wait()
	}()

	serverURL := fmt.Sprintf("http://localhost:%d", port)
	if err := awaitWebBackgroundReady(ctx, port, serverURL, child, childErrChan); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(out, "AIM Web Dashboard started in background at %s (PID %d)\n", serverURL, child.Process.Pid)
	_, _ = fmt.Fprintln(out, "To stop: aim web stop | Status: aim web status")

	if !noOpen {
		_ = openBrowserURL(serverURL)
	}

	return nil
}

func attachWebBackgroundLog(child *exec.Cmd) func() {
	baseDir := config.BaseDir()
	_ = os.MkdirAll(baseDir, 0o700)
	logPath := webLogPath()
	if fi, err := os.Stat(logPath); err == nil && fi.Size() > 5*1024*1024 {
		_ = os.Rename(logPath, logPath+".1")
	}
	root, err := os.OpenRoot(filepath.Dir(logPath))
	if err != nil {
		return func() {}
	}
	logFile, openErr := root.OpenFile(filepath.Base(logPath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if openErr != nil {
		_ = root.Close()
		return func() {}
	}
	child.Stdout = logFile
	child.Stderr = logFile
	return func() {
		_ = logFile.Close()
		_ = root.Close()
	}
}

func awaitWebBackgroundReady(ctx context.Context, port int, serverURL string, child *exec.Cmd, childErrChan <-chan error) error {
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		if checkServerHealth(ctx, port) {
			return nil
		}
		select {
		case err := <-childErrChan:
			if err != nil {
				return fmt.Errorf("web dashboard exited unexpectedly (%v). Check logs at %s", err, webLogPath())
			}
			return fmt.Errorf("web dashboard exited unexpectedly. Check logs at %s", webLogPath())
		default:
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = child.Process.Kill()
	return fmt.Errorf("timed out waiting for web dashboard to start on %s (check logs at %s)", serverURL, webLogPath())
}

func runWebForeground(cmd *cobra.Command, reg *agents.Registry, pm *profile.ProfileManager, port int, devMode, noOpen bool) error {
	if reg == nil {
		reg = defaultRegistry()
	}
	if pm == nil {
		pm = profile.NewProfileManager(config.BaseDir())
	}
	baseDir := config.BaseDir()
	if pm.BaseDir != "" {
		baseDir = pm.BaseDir
	}
	out := cmd.OutOrStdout()

	sm := getSessionManager(cmd.Context())

	launcherSvc := service.NewLauncherService()
	profileSvc := service.NewProfileService(pm, reg)
	sessionSvc := service.NewSessionServiceWithProfilesRoot(sm, pm.ProfilesRoot(), launcherSvc)
	mcpSvc := service.NewMCPService(pm, reg)
	daemonSvc := service.NewDaemonService(reg, pm, baseDir)

	shareStore := share.NewStore(baseDir)
	defer func() { _ = shareStore.Close() }()

	srv := web.NewServer(profileSvc, sessionSvc, launcherSvc, port, devMode, mcpSvc)
	srv.SetShareStore(shareStore)
	srv.SetDaemonService(daemonSvc)
	srv.SetVersion(GetVersion())

	if err := srv.Listen(); err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	actualPort := srv.Port()
	if err := writeWebPID(os.Getpid(), actualPort); err != nil {
		logger.Debug("[web] Failed to write PID file: %v", err)
	}
	defer removeOwnWebPID()

	// Start P2P Share Server listener on port 4319 in background if available
	host, _ := os.Hostname()
	shareServer := share.NewServer(shareStore, pm.ProfilesRoot(), host, 4319)
	var lc net.ListenConfig
	shareLn, shareErr := lc.Listen(cmd.Context(), "tcp", "0.0.0.0:4319")
	if shareErr == nil {
		go func() {
			if err := shareServer.Serve(shareLn); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Debug("[web] P2P Share Server stopped with error: %v", err)
			}
		}()
		defer func() {
			_ = shareServer.Close()
			_ = shareLn.Close()
		}()
		logger.Debug("[web] P2P Share Server listening on 0.0.0.0:4319")
	} else {
		logger.Debug("[web] P2P Share Server port 4319 unavailable (already running?): %v", shareErr)
	}

	serverURL := fmt.Sprintf("http://localhost:%d", actualPort)
	_, _ = fmt.Fprintf(out, "AIM Web Dashboard running at %s (PID %d)\n", serverURL, os.Getpid())
	if devMode {
		_, _ = fmt.Fprintln(out, "Development mode enabled (CORS active, Vite proxy ready)")
	}
	_, _ = fmt.Fprintln(out, "Press Ctrl+C to stop")

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if !noOpen {
		go func() {
			select {
			case <-time.After(100 * time.Millisecond):
				_ = openBrowserURL(serverURL)
			case <-ctx.Done():
			}
		}()
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Serve()
	}()

	select {
	case <-ctx.Done():
		_, _ = fmt.Fprintln(out, "\nShutting down web dashboard...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	}
}
