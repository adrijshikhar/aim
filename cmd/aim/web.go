package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	_ = os.MkdirAll(config.BaseDir(), 0755)
	return os.WriteFile(webPIDPath(), data, 0600)
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

func checkServerHealth(port int) bool {
	client := http.Client{Timeout: 400 * time.Millisecond}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/api/health", port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
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
			return startWebBackground(port, devMode, noOpen)
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
			return startWebBackground(port, devMode, noOpen)
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
			info, err := readWebPID()
			if err != nil || !isProcessAlive(info.PID) {
				_ = removeWebPID()
				fmt.Println("AIM Web Dashboard is not running.")
				return nil
			}

			if err := terminateProcess(info.PID); err != nil {
				return fmt.Errorf("failed to terminate dashboard process (PID %d): %w", info.PID, err)
			}
			_ = removeWebPID()
			fmt.Printf("🛑 AIM Web Dashboard stopped (PID %d)\n", info.PID)
			return nil
		},
	}
}

func newWebStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check the running status of the AIM web dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := readWebPID()
			if err == nil && isProcessAlive(info.PID) && checkServerHealth(info.Port) {
				fmt.Printf("● AIM Web Dashboard is running at http://localhost:%d (PID %d)\n", info.Port, info.PID)
				return nil
			}
			fmt.Println("○ AIM Web Dashboard is stopped.")
			return nil
		},
	}
}

func startWebBackground(port int, devMode, noOpen bool) error {
	// 1. Check if already running and responsive
	if info, err := readWebPID(); err == nil {
		if isProcessAlive(info.PID) && checkServerHealth(info.Port) {
			serverURL := fmt.Sprintf("http://localhost:%d", info.Port)
			fmt.Printf("🚀 AIM Web Dashboard is already running at %s (PID %d)\n", serverURL, info.PID)
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

	child := exec.Command(exe, args...)
	setSysProcAttr(child)

	_ = os.MkdirAll(config.BaseDir(), 0755)
	logPath := webLogPath()
	if fi, err := os.Stat(logPath); err == nil && fi.Size() > 5*1024*1024 {
		_ = os.Rename(logPath, logPath+".1")
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		child.Stdout = logFile
		child.Stderr = logFile
		defer logFile.Close()
	}

	if err := child.Start(); err != nil {
		return fmt.Errorf("failed to spawn web dashboard process: %w", err)
	}

	childErrChan := make(chan error, 1)
	go func() {
		childErrChan <- child.Wait()
	}()

	// 3. Poll /api/health until server is ready (up to 4 seconds)
	serverURL := fmt.Sprintf("http://localhost:%d", port)
	deadline := time.Now().Add(4 * time.Second)
	ready := false

	for time.Now().Before(deadline) {
		if checkServerHealth(port) {
			ready = true
			break
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

	if !ready {
		_ = child.Process.Kill()
		return fmt.Errorf("timed out waiting for web dashboard to start on %s (check logs at %s)", serverURL, webLogPath())
	}

	fmt.Printf("🚀 AIM Web Dashboard started in background at %s (PID %d)\n", serverURL, child.Process.Pid)
	fmt.Println("To stop: aim web stop | Status: aim web status")

	if !noOpen {
		_ = openBrowserURL(serverURL)
	}

	return nil
}

func runWebForeground(cmd *cobra.Command, reg *agents.Registry, pm *profile.ProfileManager, port int, devMode, noOpen bool) error {
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
	daemonSvc := service.NewDaemonService(reg, pm, config.BaseDir())

	srv := web.NewServer(profileSvc, sessionSvc, launcherSvc, port, devMode, mcpSvc)
	srv.SetDaemonService(daemonSvc)
	srv.SetVersion(Version)

	if err := srv.Listen(); err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	actualPort := srv.Port()
	if err := writeWebPID(os.Getpid(), actualPort); err != nil {
		logger.Debug("[web] Failed to write PID file: %v", err)
	}
	defer removeOwnWebPID()

	// Start P2P Share Server listener on port 4319 in background if available
	shareStore := share.NewStore(config.BaseDir())
	host, _ := os.Hostname()
	shareServer := share.NewServer(shareStore, pm.ProfilesRoot(), host, 4319)
	shareLn, shareErr := net.Listen("tcp", "0.0.0.0:4319")
	if shareErr == nil {
		go func() {
			_ = shareServer.Serve(shareLn)
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
	fmt.Printf("🚀 AIM Web Dashboard running at %s (PID %d)\n", serverURL, os.Getpid())
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
}
