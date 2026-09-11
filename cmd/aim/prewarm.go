package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
	"github.com/spf13/cobra"
)

func newPrewarmCmd(reg *agents.Registry, pm *profile.ProfileManager) *cobra.Command {
	return &cobra.Command{
		Use:    "__prewarm [agent]",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			agent := ""
			if len(args) > 0 {
				agent = args[0]
			}
			runPrewarm(reg, pm, agent)
		},
	}
}

func runPrewarm(reg *agents.Registry, pm *profile.ProfileManager, targetAgent string) {
	if reg == nil || pm == nil {
		return
	}
	cfg, _ := config.LoadConfig()
	var adapters []agents.AgentAdapter
	if targetAgent != "" {
		if ad, err := reg.Get(targetAgent); err == nil && ad != nil {
			adapters = []agents.AgentAdapter{ad}
		}
	} else {
		adapters = reg.All()
	}

	var targets []usage.TargetProfile
	for _, ad := range adapters {
		profs, _ := pm.ListProfilesForAgent(ad.Name(), cfg, reg)
		for _, p := range profs {
			targets = append(targets, usage.TargetProfile{
				Agent:      ad.Name(),
				Profile:    p,
				ProfileDir: pm.ProfileDir(p),
				GetUsageFn: ad.GetUsage,
			})
		}
	}

	if len(targets) == 0 {
		return
	}

	cache := usage.NewCacheStore(config.BaseDir(), usage.DefaultTTL)
	reportsChan := usage.RefreshAsync(context.Background(), targets, cache)
	for range reportsChan {
		// Draining reportsChan ensures all background probes finish
		// and are saved to cache by RefreshAsync
	}
}

func triggerPrewarmAsync(baseDir string, agent string) {
	if !usage.CanPrewarm(baseDir, time.Minute) {
		return
	}
	selfPath, err := os.Executable()
	if err != nil {
		return
	}
	// Guard against recursively spawning test binaries during test runs
	if strings.HasSuffix(selfPath, ".test") || strings.Contains(selfPath, "go-build") || strings.Contains(selfPath, "aim.test") {
		return
	}
	args := []string{"__prewarm"}
	if agent != "" {
		args = append(args, agent)
	}
	cmd := exec.Command(selfPath, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Start()
}
