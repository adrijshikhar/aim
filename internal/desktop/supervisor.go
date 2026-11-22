package desktop

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

var userDataDirRegex = regexp.MustCompile(`--user-data-dir[=\s]([^\s]+)`)

type defaultSupervisor struct {
	scanner processScanner
}

func newSupervisorWithScanner(scanner processScanner) Supervisor {
	return &defaultSupervisor{scanner: scanner}
}

func (s *defaultSupervisor) FindInstance(ctx context.Context, appName string, binaryNames []string) (*ProcessInfo, error) {
	procs, err := s.scanner.listProcesses(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing processes: %w", err)
	}

	cleanAppName := strings.TrimSuffix(strings.ToLower(appName), ".app")
	appBundleMatch := "/" + cleanAppName + ".app/"

	exactNames := make(map[string]bool)
	if cleanAppName != "" {
		exactNames[cleanAppName] = true
	}
	for _, b := range binaryNames {
		bClean := strings.TrimSuffix(strings.TrimSuffix(strings.ToLower(b), ".app"), ".exe")
		// Only consider binary names of length > 2 to ignore 1-2 char CLI aliases like "cd"
		if len(bClean) > 2 {
			exactNames[bClean] = true
		}
	}

	for _, p := range procs {
		pBinLower := strings.ToLower(filepath.ToSlash(p.binPath))
		base := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(pBinLower), ".app"), ".exe")

		// Ignore helper / auxiliary daemons located inside .app bundles
		if strings.Contains(pBinLower, "/contents/helpers/") ||
			strings.Contains(pBinLower, "/contents/frameworks/") ||
			strings.Contains(pBinLower, "/contents/resources/") {
			continue
		}

		matched := exactNames[base]
		if !matched && cleanAppName != "" && strings.Contains(pBinLower, appBundleMatch) {
			// Inside the app bundle, only accept main binaries located under Contents/MacOS
			if strings.Contains(pBinLower, "/contents/macos/") {
				matched = true
			}
		}
		if !matched {
			continue
		}

		fullCmd := strings.Join(p.args, " ")
		m := userDataDirRegex.FindStringSubmatch(fullCmd)
		var userDataDir string
		var profileName string

		if len(m) > 1 {
			userDataDir = strings.Trim(m[1], `"'`)
			parts := strings.Split(filepath.ToSlash(userDataDir), "/")
			for i, part := range parts {
				if part == "profiles" && i+1 < len(parts) {
					profileName = parts[i+1]
					break
				}
			}
		}

		return &ProcessInfo{
			PID:         p.pid,
			BinaryPath:  p.binPath,
			UserDataDir: userDataDir,
			ProfileName: profileName,
			AppName:     appName,
		}, nil
	}

	return nil, nil
}

func (s *defaultSupervisor) FocusInstance(ctx context.Context, proc *ProcessInfo) error {
	if proc == nil || proc.PID <= 0 {
		return ErrProcessNotFound
	}
	appName := proc.AppName
	if appName == "" {
		appName = strings.TrimSuffix(filepath.Base(proc.BinaryPath), ".app")
	}
	return s.scanner.activateApp(ctx, appName)
}

func (s *defaultSupervisor) GracefulQuit(ctx context.Context, proc *ProcessInfo) error {
	if proc == nil || proc.PID <= 0 {
		return ErrProcessNotFound
	}

	appName := proc.AppName
	if appName == "" {
		appName = strings.TrimSuffix(filepath.Base(proc.BinaryPath), ".app")
	}
	_ = s.scanner.quitApp(ctx, appName)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeoutCtx.Done():
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// 5-second graceful window expired; fall back to targeted SIGTERM strictly on verified PID
			_ = s.scanner.sendSignal(proc.PID, syscall.SIGTERM)
			fallbackTicker := time.NewTicker(50 * time.Millisecond)
			defer fallbackTicker.Stop()
			fallbackTimeout, fallbackCancel := context.WithTimeout(ctx, 1*time.Second)
			defer fallbackCancel()

			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-fallbackTimeout.Done():
					return ErrQuitTimeout
				case <-fallbackTicker.C:
					if !s.scanner.isAlive(proc.PID) {
						return nil
					}
				}
			}
		case <-ticker.C:
			if !s.scanner.isAlive(proc.PID) {
				return nil
			}
		}
	}
}

func (s *defaultSupervisor) EnsureSingleInstance(
	ctx context.Context,
	in io.Reader,
	out io.Writer,
	appName string,
	targetProfile, targetUserDataDir string,
	binaries []string,
	force bool,
) error {
	proc, err := s.FindInstance(ctx, appName, binaries)
	if err != nil {
		return err
	}
	if proc == nil {
		return nil
	}

	if proc.ProfileName == targetProfile || (proc.UserDataDir != "" && proc.UserDataDir == targetUserDataDir) {
		fmt.Fprintf(out, "[aim] Desktop app %s is already running for profile '%s' (PID %d). Focusing window.\n", appName, targetProfile, proc.PID)
		_ = s.FocusInstance(ctx, proc)
		return nil
	}

	activeDesc := proc.ProfileName
	if activeDesc == "" {
		activeDesc = "host default"
	}

	if !force {
		fmt.Fprintf(out, "[aim] Desktop app %s is currently running under '%s' (PID %d).\n", appName, activeDesc, proc.PID)
		fmt.Fprintf(out, "Only one profile instance may run at a time to prevent session and keychain conflicts.\n")
		fmt.Fprintf(out, "Switch to profile '%s' by closing the existing instance? [y/N]: ", targetProfile)

		reader := bufio.NewReader(in)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if strings.ToLower(line) != "y" && strings.ToLower(line) != "yes" {
			return fmt.Errorf("%w: active under '%s' (PID %d)", ErrInstanceRunning, activeDesc, proc.PID)
		}
	}

	fmt.Fprintf(out, "[aim] Gracefully closing running instance (PID %d)...\n", proc.PID)
	if err := s.GracefulQuit(ctx, proc); err != nil {
		return fmt.Errorf("closing active instance: %w", err)
	}

	fmt.Fprintf(out, "[aim] Previous instance terminated. Launching profile '%s'...\n", targetProfile)
	return nil
}
