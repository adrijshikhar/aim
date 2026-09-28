package profile

import (
	"os"
	"testing"
)

// TestMain points AIM_HOME at a scratch dir for the whole package. Several
// tests here save config or shared state without setting it themselves, and on
// a machine with a legacy ~/.aim that would overwrite the real ~/.aim/config.json.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "aim-test-home-*")
	if err != nil {
		panic(err)
	}
	_ = os.Setenv("AIM_HOME", dir)
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}
