package config

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStorageEnvChild(t *testing.T) {
	if os.Getenv("AIM_TEST_STORAGE_CHILD") != "1" {
		return
	}
	paths := []string{ConfigDir(), DataDir(), CacheDir(), StateDir(), RealHomeDir()}
	if err := json.NewEncoder(os.Stdout).Encode(paths); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestStorageEnv(t *testing.T) {
	for _, mode := range []string{"custom", "legacy", "xdg", "default"} {
		t.Run(mode, func(t *testing.T) {
			t.Cleanup(ReloadXDG)
			root := t.TempDir()
			t.Setenv("HOME", root)
			t.Setenv("AIM_REAL_HOME", "")
			t.Setenv("AIM_HOME", "")
			for _, kind := range []string{"CONFIG", "DATA", "CACHE", "STATE"} {
				t.Setenv("AIM_"+kind+"_DIR", "")
				t.Setenv("XDG_"+kind+"_HOME", "")
				if mode == "xdg" {
					t.Setenv("XDG_"+kind+"_HOME", filepath.Join(root, kind))
				}
			}
			if mode == "custom" {
				t.Setenv("AIM_HOME", filepath.Join(root, "custom store"))
			}
			if mode == "legacy" {
				if err := os.Mkdir(filepath.Join(root, ".aim"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			ReloadXDG()
			want := []string{ConfigDir(), DataDir(), CacheDir(), StateDir(), RealHomeDir()}
			env := StorageEnv()
			env["HOME"] = filepath.Join(DataDir(), "profiles", "work")
			env["AIM_TEST_STORAGE_CHILD"] = "1"
			cmd := exec.Command(os.Args[0], "-test.run=^TestStorageEnvChild$")
			for _, entry := range os.Environ() {
				key, _, _ := strings.Cut(entry, "=")
				if _, overridden := env[key]; !overridden {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			for key, value := range env {
				if strings.HasPrefix(key, "XDG_") {
					t.Fatalf("must not override provider XDG environment: %s", key)
				}
				cmd.Env = append(cmd.Env, key+"="+value)
			}
			out, err := cmd.Output()
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			if err := json.Unmarshal(out, &got); err != nil {
				t.Fatal(err)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Errorf("child path[%d] = %q; want %q", i, got[i], want[i])
				}
			}
		})
	}
}
