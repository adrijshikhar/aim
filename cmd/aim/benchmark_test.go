package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/agents/agy"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
)

func BenchmarkDispatch_Whoami_Cold(b *testing.B) {
	tempDir := b.TempDir()
	b.Setenv("AIM_HOME", tempDir)
	b.Setenv("AIM_PROFILE", "")
	b.Setenv("AIM_AGENT", "")
	b.Setenv("HOME", tempDir)
	b.Setenv("AIM_DEBUG", "")
	logger.Reset()

	reg := agents.NewRegistry()
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")
	_, _ = pm.EnsureProfile("personal")

	// Discard stdout to isolate dispatch execution latency
	nullFile, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		b.Fatal(err)
	}
	defer nullFile.Close()
	origStdout := os.Stdout
	os.Stdout = nullFile
	defer func() { os.Stdout = origStdout }()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		code := dispatch([]string{"whoami"}, reg, pm)
		if code != 0 {
			b.Fatalf("dispatch whoami failed: %d", code)
		}
	}
}

func BenchmarkDispatch_Whoami_Warm(b *testing.B) {
	tempDir := b.TempDir()
	b.Setenv("AIM_HOME", tempDir)
	b.Setenv("AIM_PROFILE", "work")
	b.Setenv("AIM_AGENT", "agy")
	profHome := filepath.Join(tempDir, "profiles", "work")
	b.Setenv("HOME", profHome)
	b.Setenv("AIM_DEBUG", "")
	logger.Reset()

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	pm := profile.NewProfileManager(tempDir)
	_, _ = pm.EnsureProfile("work")

	// Populate cache for warm whoami
	cacheStore := usage.NewCacheStore(tempDir, 1*time.Hour)
	_ = cacheStore.Put(usage.Report{
		Agent:     "agy",
		Profile:   "work",
		Status:    usage.StatusOK,
		Summary:   "5h: 85%, wk: 92%",
		FetchedAt: time.Now(),
		Windows: []usage.LimitWindow{
			{Name: "Five Hour", RemainingPct: 85, ResetsIn: 2 * time.Hour},
		},
	})

	nullFile, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		b.Fatal(err)
	}
	defer nullFile.Close()
	origStdout := os.Stdout
	os.Stdout = nullFile
	defer func() { os.Stdout = origStdout }()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		code := dispatch([]string{"whoami"}, reg, pm)
		if code != 0 {
			b.Fatalf("dispatch whoami warm failed: %d", code)
		}
	}
}

func BenchmarkDispatch_List(b *testing.B) {
	tempDir := b.TempDir()
	b.Setenv("AIM_HOME", tempDir)
	b.Setenv("AIM_DEBUG", "")
	logger.Reset()

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	pm := profile.NewProfileManager(tempDir)

	cfg := config.NewDefaultConfig()
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("bench-prof-%d", i)
		_, _ = pm.EnsureProfile(name)
		cfg.AddProfileAgent(name, "agy")
	}
	_ = config.SaveConfig(cfg)

	nullFile, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		b.Fatal(err)
	}
	defer nullFile.Close()
	origStdout := os.Stdout
	os.Stdout = nullFile
	defer func() { os.Stdout = origStdout }()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		code := dispatch([]string{"list"}, reg, pm)
		if code != 0 {
			b.Fatalf("dispatch list failed: %d", code)
		}
	}
}

func BenchmarkDispatch_List_Agent(b *testing.B) {
	tempDir := b.TempDir()
	b.Setenv("AIM_HOME", tempDir)
	b.Setenv("AIM_DEBUG", "")
	logger.Reset()

	reg := agents.NewRegistry()
	reg.Register(agy.NewAdapter())
	pm := profile.NewProfileManager(tempDir)

	cfg := config.NewDefaultConfig()
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("bench-prof-%d", i)
		_, _ = pm.EnsureProfile(name)
		cfg.AddProfileAgent(name, "agy")
	}
	_ = config.SaveConfig(cfg)

	nullFile, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		b.Fatal(err)
	}
	defer nullFile.Close()
	origStdout := os.Stdout
	os.Stdout = nullFile
	defer func() { os.Stdout = origStdout }()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		code := dispatch([]string{"list", "agy"}, reg, pm)
		if code != 0 {
			b.Fatalf("dispatch list agy failed: %d", code)
		}
	}
}
