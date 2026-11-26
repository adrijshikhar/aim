package profile

import "github.com/aim-cli/aim/internal/config"

// ShouldSeedCredentials limits explicit host credential seeding to named primary profiles.
// Single-profile setups are intentionally never auto-seeded to preserve isolation across logins/logouts.
func ShouldSeedCredentials(name string) bool {
	switch name {
	case "personal", "p", "me", "main":
		return true
	}
	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil {
		return false
	}
	return cfg.DefaultProfile == name && name != ""
}
