package main

import (
	"os"
	"testing"
)

// TestMain gives the package an empty scratch AIM_REAL_HOME. EnsureProfile links
// every top-level dotfile of the real home into a profile, so a test that
// forgets to set it would otherwise link, and could write through into, the
// user's home. A test that sets AIM_REAL_HOME itself still wins.
func TestMain(m *testing.M) {
	realHome, err := os.MkdirTemp("", "aim-test-real-home-*")
	if err != nil {
		panic(err)
	}
	_ = os.Setenv("AIM_REAL_HOME", realHome)
	code := m.Run()
	_ = os.RemoveAll(realHome)
	os.Exit(code)
}
