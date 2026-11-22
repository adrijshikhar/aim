package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/desktop"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/runner"
)

type defaultDesktopService struct {
	reg *agents.Registry
	sup desktop.Supervisor
	pm  *profile.ProfileManager
}

// NewDesktopService creates a new DesktopService with registry, supervisor, and profile manager.
func NewDesktopService(reg *agents.Registry, sup desktop.Supervisor, pm *profile.ProfileManager) DesktopService {
	return &defaultDesktopService{
		reg: reg,
		sup: sup,
		pm:  pm,
	}
}

func (s *defaultDesktopService) ListStatus(ctx context.Context) ([]DesktopStatusDTO, error) {
	var list []DesktopStatusDTO
	if s.reg == nil || s.sup == nil {
		return list, nil
	}
	for _, a := range s.reg.All() {
		da, ok := a.(agents.DesktopApp)
		if !ok || !da.IsDesktop() {
			continue
		}
		binNames := append([]string{a.BinaryName(), da.AppBundleName()}, a.Aliases()...)
		proc, _ := s.sup.FindInstance(ctx, da.AppBundleName(), binNames)

		dto := DesktopStatusDTO{
			Agent:       a.Name(),
			DisplayName: a.DisplayName(),
		}
		if proc != nil && proc.PID > 0 {
			dto.PID = proc.PID
			dto.Profile = proc.ProfileName
			dto.Running = true
		}
		list = append(list, dto)
	}
	return list, nil
}

func (s *defaultDesktopService) Launch(ctx context.Context, req DesktopLaunchRequest) error {
	if s.reg == nil {
		return fmt.Errorf("agent registry not configured")
	}
	adapter, err := s.reg.Get(req.Agent)
	if err != nil {
		return err
	}
	da, ok := adapter.(agents.DesktopApp)
	if !ok || !da.IsDesktop() {
		return fmt.Errorf("%s is not a desktop application", req.Agent)
	}

	var pDir string
	if s.pm != nil {
		var err error
		pDir, err = s.pm.EnsureProfile(req.Profile)
		if err != nil {
			return fmt.Errorf("ensuring profile: %w", err)
		}
	} else {
		pDir = filepath.Join(config.BaseDir(), "profiles", req.Profile)
	}

	launchEnv, err := adapter.PrepareEnv(ctx, req.Profile, pDir)
	if err != nil {
		return err
	}

	binNames := append([]string{adapter.BinaryName(), da.AppBundleName()}, adapter.Aliases()...)
	var in strings.Reader
	var out strings.Builder
	if err := s.sup.EnsureSingleInstance(ctx, &in, &out, da.AppBundleName(), req.Profile, da.UserDataDir(pDir), binNames, req.Force); err != nil {
		return err
	}

	cmd := exec.Command(launchEnv.BinaryPath, launchEnv.Args...)
	cmd.Dir = launchEnv.WorkingDir
	cmd.Env = runner.BuildEnv(os.Environ(), launchEnv.Env)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd.Start()
}

func (s *defaultDesktopService) Quit(ctx context.Context, agent string) error {
	if s.reg == nil {
		return fmt.Errorf("agent registry not configured")
	}
	adapter, err := s.reg.Get(agent)
	if err != nil {
		return err
	}
	da, ok := adapter.(agents.DesktopApp)
	if !ok || !da.IsDesktop() {
		return fmt.Errorf("%s is not a desktop application", agent)
	}
	binNames := append([]string{adapter.BinaryName(), da.AppBundleName()}, adapter.Aliases()...)
	proc, err := s.sup.FindInstance(ctx, da.AppBundleName(), binNames)
	if err != nil {
		return err
	}
	if proc == nil {
		return desktop.ErrProcessNotFound
	}
	return s.sup.GracefulQuit(ctx, proc)
}

func (s *defaultDesktopService) Focus(ctx context.Context, agent string) error {
	if s.reg == nil {
		return fmt.Errorf("agent registry not configured")
	}
	adapter, err := s.reg.Get(agent)
	if err != nil {
		return err
	}
	da, ok := adapter.(agents.DesktopApp)
	if !ok || !da.IsDesktop() {
		return fmt.Errorf("%s is not a desktop application", agent)
	}
	binNames := append([]string{adapter.BinaryName(), da.AppBundleName()}, adapter.Aliases()...)
	proc, err := s.sup.FindInstance(ctx, da.AppBundleName(), binNames)
	if err != nil {
		return err
	}
	if proc == nil {
		return desktop.ErrProcessNotFound
	}
	return s.sup.FocusInstance(ctx, proc)
}
