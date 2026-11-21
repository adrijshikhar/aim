package service

import (
	"os"
	"testing"
)

// TestMain provides an isolated scratch AIM_HOME and AIM_REAL_HOME for the
// internal/service package so no test profile bridges or config writes can
// touch the user's real home directory.
func TestMain(m *testing.M) {
	tempHome, err := os.MkdirTemp("", "aim-test-service-home-*")
	if err != nil {
		panic(err)
	}
	realHome, err := os.MkdirTemp("", "aim-test-service-real-home-*")
	if err != nil {
		panic(err)
	}
	_ = os.Setenv("AIM_HOME", tempHome)
	_ = os.Setenv("AIM_REAL_HOME", realHome)
	_ = os.Setenv("HOME", tempHome)

	code := m.Run()

	_ = os.RemoveAll(tempHome)
	_ = os.RemoveAll(realHome)
	os.Exit(code)
}
