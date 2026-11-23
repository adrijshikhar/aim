package agents_test

import (
	"testing"

	"github.com/aim-cli/aim/internal/agents"
)

type dummyDesktopAdapter struct {
	agents.BaseAdapter
}

func (d *dummyDesktopAdapter) IsDesktop() bool                      { return true }
func (d *dummyDesktopAdapter) AppBundleName() string                { return "Dummy" }
func (d *dummyDesktopAdapter) UserDataDir(profileDir string) string { return profileDir + "/user-data" }

func TestDesktopAppInterface(t *testing.T) {
	dummy := &dummyDesktopAdapter{
		BaseAdapter: agents.NewBaseAdapter("dummy", "Dummy", "dummy", nil),
	}
	var _ agents.DesktopApp = dummy

	if !dummy.IsDesktop() {
		t.Errorf("expected IsDesktop() to return true")
	}
	if dummy.AppBundleName() != "Dummy" {
		t.Errorf("expected AppBundleName() to return 'Dummy'")
	}
	if dummy.UserDataDir("/test") != "/test/user-data" {
		t.Errorf("expected '/test/user-data', got '%s'", dummy.UserDataDir("/test"))
	}
}
