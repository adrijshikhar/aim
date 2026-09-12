package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/aim-cli/aim/internal/logger"
)

type ProfileConfig struct {
	Agents []string          `json:"agents,omitempty"`
	Env    map[string]string `json:"env,omitempty"`
	Args   []string          `json:"args,omitempty"`
}

func (p *ProfileConfig) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		p.Agents = []string{}
		return nil
	}
	type alias ProfileConfig
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*p = ProfileConfig(a)
	return nil
}

type Config struct {
	DefaultAgent           string                   `json:"default_agent"`
	DefaultProfile         string                   `json:"default_profile"`
	CustomBridgedPaths     []string                 `json:"custom_bridged_paths,omitempty"`
	CustomIgnoredKeychains []string                 `json:"custom_ignored_keychains,omitempty"`
	Debug                  bool                     `json:"debug,omitempty"`
	Profiles               map[string]ProfileConfig `json:"profiles"`
}

func RealHomeDir() string {
	if custom := os.Getenv("AIM_REAL_HOME"); custom != "" {
		return custom
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	marker := filepath.Join(".aim", "profiles")
	if strings.Contains(home, marker) {
		idx := strings.Index(home, marker)
		if idx > 0 {
			return strings.TrimSuffix(home[:idx], string(filepath.Separator))
		}
	}
	return home
}

func BaseDir() string {
	if custom := os.Getenv("AIM_HOME"); custom != "" {
		return custom
	}
	return filepath.Join(RealHomeDir(), ".aim")
}

func ConfigFilePath() string {
	return filepath.Join(BaseDir(), "config.json")
}

func NewDefaultConfig() *Config {
	return &Config{
		DefaultAgent:   "agy",
		DefaultProfile: "default",
		Profiles:       make(map[string]ProfileConfig),
	}
}

func (c *Config) GetProfileAgents(profile string) []string {
	if c == nil || c.Profiles == nil {
		return nil
	}
	p, ok := c.Profiles[profile]
	if !ok {
		return nil
	}
	return append([]string(nil), p.Agents...)
}

func (c *Config) AddProfileAgent(profile, agent string) {
	if c == nil {
		return
	}
	if c.Profiles == nil {
		c.Profiles = make(map[string]ProfileConfig)
	}
	p := c.Profiles[profile]
	for _, a := range p.Agents {
		if a == agent {
			return
		}
	}
	p.Agents = append(p.Agents, agent)
	c.Profiles[profile] = p
}

func (c *Config) RemoveProfileAgent(profile, agent string) {
	if c == nil || c.Profiles == nil {
		return
	}
	p, ok := c.Profiles[profile]
	if !ok {
		return
	}
	var next []string
	for _, a := range p.Agents {
		if a != agent {
			next = append(next, a)
		}
	}
	p.Agents = next
	c.Profiles[profile] = p
}

func (c *Config) HasAgent(profile, agent string) bool {
	if c == nil {
		return false
	}
	for _, a := range c.GetProfileAgents(profile) {
		if a == agent {
			return true
		}
	}
	return false
}

func (c *Config) DeleteProfile(profile string) {
	if c == nil || c.Profiles == nil {
		return
	}
	delete(c.Profiles, profile)
}

func (c *Config) RenameProfile(oldProfile, newProfile string) {
	if c == nil || c.Profiles == nil {
		return
	}
	p, ok := c.Profiles[oldProfile]
	if !ok {
		return
	}
	delete(c.Profiles, oldProfile)
	c.Profiles[newProfile] = p
	if c.DefaultProfile == oldProfile {
		c.DefaultProfile = newProfile
	}
}

func (c *Config) GetProfileEnv(profile string) map[string]string {
	if c == nil || c.Profiles == nil {
		return nil
	}
	p, ok := c.Profiles[profile]
	if !ok || p.Env == nil {
		return nil
	}
	out := make(map[string]string, len(p.Env))
	for k, v := range p.Env {
		out[k] = v
	}
	return out
}

func (c *Config) SetProfileEnv(profile string, env map[string]string) {
	if c == nil {
		return
	}
	if c.Profiles == nil {
		c.Profiles = make(map[string]ProfileConfig)
	}
	p := c.Profiles[profile]
	if env == nil {
		p.Env = nil
	} else {
		p.Env = make(map[string]string, len(env))
		for k, v := range env {
			p.Env[k] = v
		}
	}
	c.Profiles[profile] = p
}

func (c *Config) GetProfileArgs(profile string) []string {
	if c == nil || c.Profiles == nil {
		return nil
	}
	p, ok := c.Profiles[profile]
	if !ok || p.Args == nil {
		return nil
	}
	return append([]string(nil), p.Args...)
}

func (c *Config) SetProfileArgs(profile string, args []string) {
	if c == nil {
		return
	}
	if c.Profiles == nil {
		c.Profiles = make(map[string]ProfileConfig)
	}
	p := c.Profiles[profile]
	if args == nil {
		p.Args = nil
	} else {
		p.Args = append([]string(nil), args...)
	}
	c.Profiles[profile] = p
}

func LoadConfig() (*Config, error) {
	path := ConfigFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewDefaultConfig(), nil
		}
		return nil, err
	}
	cfg := NewDefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]ProfileConfig)
	}
	if cfg.Debug {
		env := strings.TrimSpace(strings.ToLower(os.Getenv("AIM_DEBUG")))
		if env != "0" && env != "false" && env != "no" && env != "off" {
			logger.SetDebug(true)
		}
	}
	return cfg, nil
}

func SaveConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("cannot save nil config")
	}
	base := BaseDir()
	if err := os.MkdirAll(base, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmpFile := filepath.Join(base, "config.json.tmp")
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmpFile, ConfigFilePath())
}
