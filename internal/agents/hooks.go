package agents

import (
	"fmt"
	"os"
	"path/filepath"
)

// SetupLeasedProfileHooks provisions isolated, agent-native guard hooks for a leased profile.
// It ensures host hooks, host hooks.json, and host plugins are never symlinked or executed,
// and installs the aim hook guard for prompt/tool execution.
func SetupLeasedProfileHooks(agentName, profileName, profileDir string) error {
	return setupLeasedProfileHooks(agentName, profileName, profileDir)
}

func setupLeasedProfileHooks(agentName, profileName, profileDir string) error {
	switch agentName {
	case "codex":
		codexDir := filepath.Join(profileDir, ".codex")
		if err := os.MkdirAll(codexDir, 0700); err != nil {
			return err
		}
		// Remove any existing symlinks to host hooks or plugins
		for _, name := range []string{"plugins", "hooks", "hooks.json"} {
			p := filepath.Join(codexDir, name)
			if fi, err := os.Lstat(p); err == nil && fi.Mode()&os.ModeSymlink != 0 {
				_ = os.Remove(p)
			}
		}
		hooksPath := filepath.Join(codexDir, "hooks.json")
		isolatedHooks := fmt.Sprintf(`{
  "hooks": {
    "UserPromptSubmit": [
      { "hooks": [ { "type": "command", "command": "aim hook guard %s --agent codex" } ] }
    ],
    "PreToolUse": [
      { "hooks": [ { "type": "command", "command": "aim hook guard %s --agent codex" } ] }
    ],
    "SubagentStart": [
      { "hooks": [ { "type": "command", "command": "aim hook guard %s --agent codex" } ] }
    ]
  }
}`, profileName, profileName, profileName)
		return os.WriteFile(hooksPath, []byte(isolatedHooks), 0600)

	case "claude":
		destDir := filepath.Join(profileDir, ".claude")
		if err := os.MkdirAll(destDir, 0700); err != nil {
			return err
		}
		destSettings := filepath.Join(destDir, "settings.json")
		if fi, err := os.Lstat(destSettings); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			_ = os.Remove(destSettings)
		}
		isolatedSettings := fmt.Sprintf(`{
  "hooks": {
    "UserPromptSubmit": [
      { "command": "aim hook guard %s --agent claude", "type": "command" }
    ],
    "PreToolUse": [
      { "command": "aim hook guard %s --agent claude", "type": "command" }
    ]
  }
}`, profileName, profileName)
		return os.WriteFile(destSettings, []byte(isolatedSettings), 0600)

	case "agy":
		profileConfigDir := filepath.Join(profileDir, ".gemini", "config")
		if err := os.MkdirAll(profileConfigDir, 0700); err != nil {
			return err
		}
		// Remove any existing symlinks to host hooks or plugins
		for _, name := range []string{"plugins", "hooks", "hooks.json"} {
			p := filepath.Join(profileConfigDir, name)
			if fi, err := os.Lstat(p); err == nil && fi.Mode()&os.ModeSymlink != 0 {
				_ = os.Remove(p)
			}
		}
		hooksPath := filepath.Join(profileConfigDir, "hooks.json")
		isolatedHooks := fmt.Sprintf(`{
  "aim-guard": {
    "PreInvocation": [
      { "type": "command", "command": "aim hook guard %s --agent agy", "timeout": 10 }
    ],
    "PreToolUse": [
      { "type": "command", "command": "aim hook guard %s --agent agy", "timeout": 10 }
    ]
  }
}`, profileName, profileName)
		return os.WriteFile(hooksPath, []byte(isolatedHooks), 0600)

	default:
		return nil
	}
}
