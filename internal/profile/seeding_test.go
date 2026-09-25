package profile

import (
	"os"
	"testing"

	"github.com/aim-cli/aim/internal/config"
)

func TestShouldSeedCredentials(t *testing.T) {
	for _, tt := range []struct {
		name, profile, config string
		want                  bool
	}{
		{"personal alias", "personal", `{}`, true},
		{"short alias", "p", `{}`, true},
		{"me alias", "me", `{}`, true},
		{"main alias", "main", `{}`, true},
		{"default", "work", `{"default_profile":"work","profiles":{"work":{},"other":{}}}`, true},
		{"sole profile", "work", `{"profiles":{"work":{}}}`, true},
		{"other sole profile", "work", `{"profiles":{"other":{}}}`, false},
		{"multiple profiles", "work", `{"profiles":{"work":{},"other":{}}}`, false},
		{"empty config", "work", `{}`, false},
		{"malformed config", "work", `{`, false},
		{"alias ignores malformed config", "personal", `{`, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AIM_HOME", t.TempDir())
			if err := os.WriteFile(config.ConfigFilePath(), []byte(tt.config), 0600); err != nil {
				t.Fatal(err)
			}
			if got := ShouldSeedCredentials(tt.profile); got != tt.want {
				t.Errorf("ShouldSeedCredentials(%q) = %v; want %v", tt.profile, got, tt.want)
			}
		})
	}
}
