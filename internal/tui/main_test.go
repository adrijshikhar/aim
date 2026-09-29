package tui

import (
	"os"
	"testing"
)

// TestMain points AIM_HOME at a scratch dir for the whole package. Several
// tests here save config or shared state without setting it themselves, and on
// a machine with a legacy ~/.aim that would overwrite the real ~/.aim/config.json.
// AIM_REAL_HOME gets its own empty scratch dir too: EnsureProfile links every
// top-level dotfile of the real home into a profile, so a test that forgets to
// set it would otherwise link, and could write through into, the user's home.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "aim-test-home-*")
	if err != nil {
		panic(err)
	}
	realHome, err := os.MkdirTemp("", "aim-test-real-home-*")
	if err != nil {
		panic(err)
	}
	_ = os.Setenv("AIM_HOME", dir)
	_ = os.Setenv("AIM_REAL_HOME", realHome)
	code := m.Run()
	_ = os.RemoveAll(dir)
	_ = os.RemoveAll(realHome)
	os.Exit(code)
}
