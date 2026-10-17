package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

type ProfileConfig struct {
	Agents        []string          `json:"agents,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Args          []string          `json:"args,omitempty"`
	MCPGlobal     *bool             `json:"mcp_global,omitempty"`
	PluginsGlobal *bool             `json:"plugins_global,omitempty"`
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
	TelemetryDisabled      bool                     `json:"telemetry_disabled,omitempty"`
	AutoUpdate             string                   `json:"auto_update,omitempty"`
	Profiles               map[string]ProfileConfig `json:"profiles"`
}

func (c *Config) UnmarshalJSON(data []byte) error {
	type alias Config
	raw := struct {
		alias
		TelemetryDisabled *bool `json:"telemetry_disabled,omitempty"`
		Telemetry         *any  `json:"telemetry,omitempty"`
	}{
		alias: alias(*c),
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*c = Config(raw.alias)

	if raw.TelemetryDisabled != nil {
		c.TelemetryDisabled = *raw.TelemetryDisabled
	} else if raw.Telemetry != nil {
		switch v := (*raw.Telemetry).(type) {
		case bool:
			c.TelemetryDisabled = !v
		case string:
			lower := strings.ToLower(strings.TrimSpace(v))
			if lower == "off" || lower == "false" || lower == "0" || lower == "disabled" {
				c.TelemetryDisabled = true
			} else if lower == "on" || lower == "true" || lower == "1" || lower == "enabled" {
				c.TelemetryDisabled = false
			}
		}
	}
	return nil
}

// Get returns the string representation of a known configuration key.
func (c *Config) Get(key string) (string, error) {
	if c == nil {
		return "", errors.New("cannot read from nil config")
	}
	normKey := strings.ToLower(strings.TrimSpace(key))
	switch normKey {
	case "default_agent":
		return c.DefaultAgent, nil
	case "default_profile":
		return c.DefaultProfile, nil
	case "auto_update":
		return c.GetAutoUpdate(), nil
	case "debug":
		if c.Debug {
			return "true", nil
		}
		return "false", nil
	case "telemetry":
		if c.TelemetryDisabled {
			return "off", nil
		}
		return "on", nil
	case "telemetry_disabled":
		if c.TelemetryDisabled {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("unknown config key %q (supported: default_agent, default_profile, auto_update, telemetry, telemetry_disabled, debug)", key)
	}
}

// Set mutates a known configuration key after validating the value.
func (c *Config) Set(key, val string) error {
	if c == nil {
		return errors.New("cannot write to nil config")
	}
	val = strings.TrimSpace(val)
	normKey := strings.ToLower(strings.TrimSpace(key))
	switch normKey {
	case "default_agent":
		if val == "" {
			return errors.New("default_agent cannot be empty")
		}
		c.DefaultAgent = strings.ToLower(val)
		return nil
	case "default_profile":
		c.DefaultProfile = val
		return nil
	case "auto_update":
		valLower := strings.ToLower(val)
		switch valLower {
		case AutoUpdatePrompt, AutoUpdateNotify, AutoUpdateAuto, AutoUpdateOff:
			c.AutoUpdate = valLower
			return nil
		case "false", "0", "disabled":
			c.AutoUpdate = AutoUpdateOff
			return nil
		case "true", "1", "enabled":
			c.AutoUpdate = AutoUpdateAuto
			return nil
		default:
			return fmt.Errorf("invalid auto_update mode %q (must be one of: prompt, notify, auto, off)", val)
		}
	case "debug":
		b, err := parseBool(val)
		if err != nil {
			return fmt.Errorf("invalid boolean for debug %q: must be true or false", val)
		}
		c.Debug = b
		return nil
	case "telemetry":
		valLower := strings.ToLower(val)
		switch valLower {
		case "off", "false", "0", "disabled":
			c.TelemetryDisabled = true
			return nil
		case "on", "true", "1", "enabled":
			c.TelemetryDisabled = false
			return nil
		default:
			return fmt.Errorf("invalid telemetry setting %q (must be 'on' or 'off')", val)
		}
	case "telemetry_disabled":
		b, err := parseBool(val)
		if err != nil {
			return fmt.Errorf("invalid boolean for telemetry_disabled %q: must be true or false", val)
		}
		c.TelemetryDisabled = b
		return nil
	default:
		return fmt.Errorf("unknown config key %q (supported: default_agent, default_profile, auto_update, telemetry, telemetry_disabled, debug)", key)
	}
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "on", "enabled":
		return true, nil
	case "false", "0", "no", "off", "disabled":
		return false, nil
	default:
		return false, fmt.Errorf("cannot parse %q as boolean", s)
	}
}

const (
	AutoUpdatePrompt = "prompt"
	AutoUpdateNotify = "notify"
	AutoUpdateAuto   = "auto"
	AutoUpdateOff    = "off"
)

func RealHomeDir() string {
	if custom := os.Getenv("AIM_REAL_HOME"); custom != "" {
		return custom
	}
	home := os.Getenv("HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "."
		}
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

func isLegacy() bool {
	if custom := os.Getenv("AIM_HOME"); custom != "" {
		return true
	}
	legacyDir := filepath.Join(RealHomeDir(), ".aim")
	_, err := os.Stat(legacyDir)
	return err == nil
}

func legacyBaseDir() string {
	if custom := os.Getenv("AIM_HOME"); custom != "" {
		return custom
	}
	return filepath.Join(RealHomeDir(), ".aim")
}

// ConfigDir returns the configuration directory:
// $XDG_CONFIG_HOME/aim or ~/.config/aim (fallback ~/.aim if legacy installation or AIM_HOME is set).
func ConfigDir() string {
	if isLegacy() {
		return legacyBaseDir()
	}
	if dir := os.Getenv("AIM_CONFIG_DIR"); dir != "" {
		return dir
	}
	if custom := os.Getenv("XDG_CONFIG_HOME"); custom != "" {
		p := filepath.Join(custom, "aim")
		_ = os.MkdirAll(p, 0700)
		return p
	}
	if p, err := xdg.ConfigFile(filepath.Join("aim", "config.json")); err == nil {
		return filepath.Dir(p)
	}
	return filepath.Join(xdg.ConfigHome, "aim")
}

// DataDir returns the data directory for profiles and state:
// $XDG_DATA_HOME/aim or ~/.local/share/aim (fallback ~/.aim if legacy installation or AIM_HOME is set).
func DataDir() string {
	if isLegacy() {
		return legacyBaseDir()
	}
	if dir := os.Getenv("AIM_DATA_DIR"); dir != "" {
		return dir
	}
	if custom := os.Getenv("XDG_DATA_HOME"); custom != "" {
		p := filepath.Join(custom, "aim")
		_ = os.MkdirAll(p, 0700)
		return p
	}
	if p, err := xdg.DataFile(filepath.Join("aim", "profiles")); err == nil {
		return filepath.Dir(p)
	}
	return filepath.Join(xdg.DataHome, "aim")
}

// CacheDir returns the cache directory:
// $XDG_CACHE_HOME/aim or ~/.cache/aim (fallback ~/.aim/cache if legacy installation or AIM_HOME is set).
func CacheDir() string {
	if isLegacy() {
		return filepath.Join(legacyBaseDir(), "cache")
	}
	if dir := os.Getenv("AIM_CACHE_DIR"); dir != "" {
		return dir
	}
	if custom := os.Getenv("XDG_CACHE_HOME"); custom != "" {
		p := filepath.Join(custom, "aim")
		_ = os.MkdirAll(p, 0700)
		return p
	}
	if p, err := xdg.CacheFile(filepath.Join("aim", "cache.lock")); err == nil {
		dir := filepath.Dir(p)
		_ = os.MkdirAll(dir, 0700)
		return dir
	}
	dir := filepath.Join(xdg.CacheHome, "aim")
	_ = os.MkdirAll(dir, 0700)
	return dir
}

// StateDir returns the state/logs directory:
// $XDG_STATE_HOME/aim or ~/.local/state/aim (fallback ~/.aim if legacy installation or AIM_HOME is set).
func StateDir() string {
	if isLegacy() {
		return legacyBaseDir()
	}
	if dir := os.Getenv("AIM_STATE_DIR"); dir != "" {
		return dir
	}
	if custom := os.Getenv("XDG_STATE_HOME"); custom != "" {
		p := filepath.Join(custom, "aim")
		_ = os.MkdirAll(p, 0700)
		return p
	}
	if p, err := xdg.StateFile(filepath.Join("aim", "aim-debug.log")); err == nil {
		return filepath.Dir(p)
	}
	return filepath.Join(xdg.StateHome, "aim")
}

// BaseDir returns the backward-compatible base directory prioritizing ~/.aim if it exists or AIM_HOME is set.
// On fresh installations adhering to XDG, it returns DataDir().
func BaseDir() string {
	if isLegacy() {
		return legacyBaseDir()
	}
	return DataDir()
}

// StorageEnv preserves AIM's resolved storage paths when a child gets a profile
// HOME. AIM-specific overrides leave the provider's XDG defaults isolated.
// AIM_HOME remains a legacy-layout selector, not an alias for DataDir.
func StorageEnv() map[string]string {
	env := map[string]string{
		"AIM_HOME":       "",
		"AIM_REAL_HOME":  RealHomeDir(),
		"AIM_CONFIG_DIR": ConfigDir(),
		"AIM_DATA_DIR":   DataDir(),
		"AIM_CACHE_DIR":  CacheDir(),
		"AIM_STATE_DIR":  StateDir(),
	}
	if isLegacy() {
		env["AIM_HOME"] = legacyBaseDir()
	}
	return env
}

// ConfigFilePath returns the path to config.json within ConfigDir().
func ConfigFilePath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// ReloadXDG reloads XDG environment variables. Useful for tests.
func ReloadXDG() {
	xdg.Reload()
}

func NewDefaultConfig() *Config {
	return &Config{
		DefaultAgent:   "agy",
		DefaultProfile: "",
		AutoUpdate:     AutoUpdatePrompt,
		Profiles:       make(map[string]ProfileConfig),
	}
}

// GetAutoUpdate returns the active auto_update mode.
// Priority: AIM_AUTO_UPDATE environment variable -> cfg.AutoUpdate -> AutoUpdatePrompt ("prompt").
func (c *Config) GetAutoUpdate() string {
	if env := strings.TrimSpace(strings.ToLower(os.Getenv("AIM_AUTO_UPDATE"))); env != "" {
		switch env {
		case AutoUpdatePrompt, AutoUpdateNotify, AutoUpdateAuto, AutoUpdateOff:
			return env
		case "false", "0", "disabled":
			return AutoUpdateOff
		case "true", "1", "enabled":
			return AutoUpdateAuto
		}
	}
	if c != nil {
		mode := strings.TrimSpace(strings.ToLower(c.AutoUpdate))
		switch mode {
		case AutoUpdatePrompt, AutoUpdateNotify, AutoUpdateAuto, AutoUpdateOff:
			return mode
		}
	}
	return AutoUpdatePrompt
}

// MCPGlobalEnabled reports whether host MCP servers are merged into the
// profile's sessions. Absent means true.
func (c *Config) MCPGlobalEnabled(profile string) bool {
	if c == nil {
		return true
	}
	p, ok := c.Profiles[profile]
	if !ok || p.MCPGlobal == nil {
		return true
	}
	return *p.MCPGlobal
}

// PluginsGlobalEnabled reports whether host plugin enablement is merged into
// the profile's sessions. Absent means true.
func (c *Config) PluginsGlobalEnabled(profile string) bool {
	if c == nil {
		return true
	}
	p, ok := c.Profiles[profile]
	if !ok || p.PluginsGlobal == nil {
		return true
	}
	return *p.PluginsGlobal
}

// GroupEnabled reports whether the host merge of a collection group is on for
// the profile: "mcp" follows mcp_global, "plugins" plugins_global. Any other
// group has no setting and is always merged.
func (c *Config) GroupEnabled(profile, group string) bool {
	switch group {
	case "mcp":
		return c.MCPGlobalEnabled(profile)
	case "plugins":
		return c.PluginsGlobalEnabled(profile)
	}
	return true
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
	if c == nil || c.Profiles == nil {
		return false
	}
	p, ok := c.Profiles[profile]
	if !ok {
		return false
	}
	for _, a := range p.Agents {
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
	if c.DefaultProfile == profile {
		c.DefaultProfile = ""
	}
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
	return maps.Clone(p.Env)
}

func (c *Config) SetProfileEnv(profile string, env map[string]string) {
	if c == nil {
		return
	}
	if c.Profiles == nil {
		c.Profiles = make(map[string]ProfileConfig)
	}
	p := c.Profiles[profile]
	p.Env = maps.Clone(env)
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
	if len(cfg.Profiles) == 0 && cfg.DefaultProfile == "default" {
		cfg.DefaultProfile = ""
	}
	return cfg, nil
}

func SaveConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("cannot save nil config")
	}
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "config.*.json.tmp")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync temp config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp config: %w", err)
	}

	return os.Rename(tmpName, ConfigFilePath())
}
