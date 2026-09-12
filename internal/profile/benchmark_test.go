package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
)

func setupMockHostHome(t testing.TB) string {
	t.Helper()
	homeDir := t.TempDir()

	// Create typical dotfiles in homeDir
	dotfiles := []string{
		".gitconfig",
		".zshrc",
		".bashrc",
		".ssh/config",
		".npmrc",
		".cargo/config.toml",
		".aws/credentials",
		".kube/config",
	}
	for _, df := range dotfiles {
		p := filepath.Join(homeDir, df)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		if err := os.WriteFile(p, []byte("# mock dotfile content\n"), 0644); err != nil {
			t.Fatalf("write file failed: %v", err)
		}
	}
	return homeDir
}

func BenchmarkResolveProfile(b *testing.B) {
	tmpDir := b.TempDir()
	pm := NewProfileManager(tmpDir)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		pDir, err := pm.ResolveProfile("production")
		if err != nil || pDir == "" {
			b.Fatalf("ResolveProfile failed: %v", err)
		}
	}
}

func BenchmarkEnsureProfile(b *testing.B) {
	mockHome := setupMockHostHome(b)
	tmpDir := b.TempDir()
	b.Setenv("AIM_HOME", tmpDir)
	b.Setenv("AIM_REAL_HOME", mockHome)
	b.Setenv("AIM_DEBUG", "")
	logger.Reset()

	pm := NewProfileManager(tmpDir)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("bench-prof-%d", i%50)
		pDir, err := pm.EnsureProfile(name)
		if err != nil || pDir == "" {
			b.Fatalf("EnsureProfile failed: %v", err)
		}
	}
}

func BenchmarkEnsureDotfiles_Fresh(b *testing.B) {
	mockHome := setupMockHostHome(b)
	tmpBase := b.TempDir()
	b.Setenv("AIM_DEBUG", "")
	logger.Reset()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		profDir := filepath.Join(tmpBase, fmt.Sprintf("fresh-%d", i))
		if err := os.MkdirAll(profDir, 0700); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		if err := EnsureDotfiles(mockHome, profDir); err != nil {
			b.Fatalf("EnsureDotfiles fresh failed: %v", err)
		}
	}
}

func BenchmarkEnsureDotfiles_Warm(b *testing.B) {
	mockHome := setupMockHostHome(b)
	profDir := filepath.Join(b.TempDir(), "warm-prof")
	b.Setenv("AIM_DEBUG", "")
	logger.Reset()

	if err := os.MkdirAll(profDir, 0700); err != nil {
		b.Fatal(err)
	}
	if err := EnsureDotfiles(mockHome, profDir); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := EnsureDotfiles(mockHome, profDir); err != nil {
			b.Fatalf("EnsureDotfiles warm failed: %v", err)
		}
	}
}

func BenchmarkListProfiles(b *testing.B) {
	tmpDir := b.TempDir()
	pm := NewProfileManager(tmpDir)
	for i := 0; i < 20; i++ {
		pDir := filepath.Join(pm.ProfilesRoot(), fmt.Sprintf("profile-%02d", i))
		_ = os.MkdirAll(pDir, 0700)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		profs, err := pm.ListProfiles()
		if err != nil || len(profs) != 20 {
			b.Fatalf("ListProfiles failed: %v (len=%d)", err, len(profs))
		}
	}
}

func BenchmarkListProfilesForAgent(b *testing.B) {
	tmpDir := b.TempDir()
	pm := NewProfileManager(tmpDir)
	cfg := config.NewDefaultConfig()

	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("profile-%02d", i)
		pDir := filepath.Join(pm.ProfilesRoot(), name)
		_ = os.MkdirAll(pDir, 0700)
		if i%2 == 0 {
			cfg.AddProfileAgent(name, "agy")
		} else {
			cfg.AddProfileAgent(name, "gemini")
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		profs, err := pm.ListProfilesForAgent("agy", cfg, nil)
		if err != nil || len(profs) != 10 {
			b.Fatalf("ListProfilesForAgent failed: %v (len=%d)", err, len(profs))
		}
	}
}

func BenchmarkCloneProfile(b *testing.B) {
	mockHome := setupMockHostHome(b)
	tmpDir := b.TempDir()
	b.Setenv("AIM_HOME", tmpDir)
	b.Setenv("AIM_REAL_HOME", mockHome)
	b.Setenv("AIM_DEBUG", "")
	logger.Reset()

	pm := NewProfileManager(tmpDir)
	srcDir := filepath.Join(pm.ProfilesRoot(), "source-prof")
	_ = os.MkdirAll(srcDir, 0700)
	// Add dummy files and dirs in source
	_ = os.WriteFile(filepath.Join(srcDir, "config.json"), []byte(`{"theme": "dark"}`), 0644)
	_ = os.WriteFile(filepath.Join(srcDir, "custom.env"), []byte("FOO=bar\nBAZ=qux\n"), 0644)
	_ = os.MkdirAll(filepath.Join(srcDir, "sub"), 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "sub", "data.txt"), []byte("sample sub data"), 0644)

	cfg := config.NewDefaultConfig()
	cfg.AddProfileAgent("source-prof", "agy")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		dstName := fmt.Sprintf("clone-%d", i)
		b.StartTimer()

		if err := pm.CloneProfile("source-prof", dstName, "agy", cfg); err != nil {
			b.Fatalf("CloneProfile failed: %v", err)
		}
	}
}
