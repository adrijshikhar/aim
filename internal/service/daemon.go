package service

import (
	"context"
	"fmt"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/daemon"
	"github.com/aim-cli/aim/internal/profile"
)

type daemonService struct {
	reg     *agents.Registry
	pm      *profile.ProfileManager
	baseDir string
}

// NewDaemonService constructs a DaemonService wrapping the internal daemon package.
func NewDaemonService(reg *agents.Registry, pm *profile.ProfileManager, baseDir string) DaemonService {
	return &daemonService{
		reg:     reg,
		pm:      pm,
		baseDir: baseDir,
	}
}

func (s *daemonService) GetStatus(ctx context.Context) (*DaemonDTO, error) {
	info, err := daemon.Status(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve daemon status: %w", err)
	}
	return toDaemonDTO(info), nil
}

func (s *daemonService) Install(ctx context.Context, binaryPath string) (*DaemonDTO, error) {
	info, err := daemon.Install(binaryPath, s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to install daemon: %w", err)
	}
	return toDaemonDTO(info), nil
}

func (s *daemonService) Uninstall(ctx context.Context) error {
	if err := daemon.Uninstall(s.baseDir); err != nil {
		return fmt.Errorf("failed to uninstall daemon: %w", err)
	}
	return nil
}

func (s *daemonService) RunOnce(ctx context.Context) error {
	if err := daemon.RunOnce(ctx, s.reg, s.pm, s.baseDir); err != nil {
		return fmt.Errorf("failed to execute daemon run: %w", err)
	}
	return nil
}

func toDaemonDTO(info *daemon.ServiceInfo) *DaemonDTO {
	if info == nil {
		return &DaemonDTO{}
	}
	sec := int(info.Interval.Seconds())
	if sec <= 0 && (info.Installed || info.Active) {
		sec = 900
	}
	return &DaemonDTO{
		Installed:      info.Installed,
		Active:         info.Active,
		Label:          info.Label,
		ConfigPath:     info.ConfigPath,
		IntervalSec:    sec,
		BinaryPath:     info.BinaryPath,
		LogPath:        info.LogPath,
		LastRun:        info.LastRun,
		LastRunMessage: info.LastRunMessage,
	}
}
