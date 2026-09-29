package tui

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/aim-cli/aim/internal/usage"
	tea "github.com/charmbracelet/bubbletea"
)

// doctorRecorder records the profile dirs Doctor is called with.
type doctorRecorder struct{ dirs []string }

func (d *doctorRecorder) Name() string                          { return "agy" }
func (d *doctorRecorder) DisplayName() string                   { return "agy" }
func (d *doctorRecorder) Aliases() []string                     { return nil }
func (d *doctorRecorder) BinaryName() string                    { return "agy" }
func (d *doctorRecorder) HasCredentials(profileDir string) bool { return false }
func (d *doctorRecorder) Login(ctx context.Context, profileName, profileDir string) error {
	return nil
}
func (d *doctorRecorder) PrepareEnv(profileName, profileDir string) (agents.LaunchEnv, error) {
	return agents.LaunchEnv{}, nil
}
func (d *doctorRecorder) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	d.dirs = append(d.dirs, profileDir)
	return nil
}
func (d *doctorRecorder) GetUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	return &usage.Report{}, nil
}

func TestDoctorDrawer_NoProfilesDoesNotCallDoctorWithEmptyDir(t *testing.T) {
	base := t.TempDir()
	t.Setenv("AIM_HOME", base)
	rec := &doctorRecorder{}
	reg := agents.NewRegistry()
	reg.Register(rec)
	m := NewModel(reg, profile.NewProfileManager(base), config.NewDefaultConfig())
	opened, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !opened.(Model).IsDoctorDrawerActive() {
		t.Fatal("drawer must open with no profiles")
	}
	for _, d := range rec.dirs {
		if d == "" {
			t.Fatal(`Doctor was called with profileDir ""`)
		}
	}
}

func TestDoctorDrawer_ShowsBridgeResults(t *testing.T) {
	base := t.TempDir()
	host := t.TempDir()
	t.Setenv("AIM_HOME", base)
	t.Setenv("AIM_REAL_HOME", host)
	if err := os.WriteFile(filepath.Join(host, ".gitconfig"), []byte("[user]"), 0644); err != nil {
		t.Fatal(err)
	}
	pm := profile.NewProfileManager(base)
	if _, err := pm.EnsureProfile("p"); err != nil {
		t.Fatal(err)
	}
	cfg := config.NewDefaultConfig()
	cfg.Profiles = map[string]config.ProfileConfig{"p": {Agents: []string{"agy"}}}
	reg := agents.NewRegistry()
	reg.Register(&doctorRecorder{})
	m := NewModel(reg, pm, cfg)
	opened, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	found := false
	for _, r := range opened.(Model).DoctorDrawerResults() {
		if r.Category == "Bridge" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a Bridge result in the drawer, got %+v", opened.(Model).DoctorDrawerResults())
	}
}
