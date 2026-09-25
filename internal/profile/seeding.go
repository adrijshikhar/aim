package profile

import "github.com/aim-cli/aim/internal/config"

// ShouldSeedCredentials limits host credential seeding to the user's primary profile.
func ShouldSeedCredentials(name string) bool {
	switch name {
	case "personal", "p", "me", "main":
		return true
	}
	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil {
		return false
	}
	if cfg.DefaultProfile == name && name != "" {
		return true
	}
	_, exists := cfg.Profiles[name]
	return len(cfg.Profiles) == 1 && exists
}
