package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/service"
	"github.com/aim-cli/aim/internal/web"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var openBrowserURL = browser.OpenURL

func newWebCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	var (
		port    int
		devMode bool
		noOpen  bool
	)

	cmd := &cobra.Command{
		Use:   "web [flags]",
		Short: "Start the AIM web dashboard",
		Long: `Start the AIM web dashboard HTTP server with REST APIs and embedded UI.

Examples:
  aim web
  aim web --port 8080
  aim web --port 3000 --dev
  aim web --no-open`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if reg == nil {
				reg = defaultRegistry()
			}
			if pm == nil {
				pm = profile.NewProfileManager(config.BaseDir())
			}

			sm := getSessionManager(cmd.Context())

			launcherSvc := service.NewLauncherService()
			profileSvc := service.NewProfileService(pm, reg)
			sessionSvc := service.NewSessionService(sm, launcherSvc)
			mcpSvc := service.NewMCPService(pm)

			srv := web.NewServer(profileSvc, sessionSvc, launcherSvc, port, devMode, mcpSvc)
			srv.SetVersion(Version)

			if err := srv.Listen(); err != nil {
				return fmt.Errorf("failed to listen on port %d: %w", port, err)
			}

			serverURL := fmt.Sprintf("http://localhost:%d", srv.Port())
			fmt.Printf("🚀 AIM Web Dashboard running at %s\n", serverURL)
			if devMode {
				fmt.Println("🛠️  Development mode enabled (CORS active, Vite proxy ready)")
			}
			fmt.Println("Press Ctrl+C to stop")

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
				fmt.Println("\nShutting down web dashboard...")
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer shutdownCancel()
				return srv.Shutdown(shutdownCtx)
			case err := <-serverErr:
				if err != nil {
					return fmt.Errorf("server error: %w", err)
				}
				return nil
			}
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	cmd.Flags().BoolVar(&devMode, "dev", false, "Run in development mode (CORS and proxy to Vite dev server)")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open browser automatically")

	return cmd
}
