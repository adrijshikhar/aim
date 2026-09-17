---
name: onboard-provider
description: Use when onboarding, implementing, refactoring, or testing an AI CLI agent adapter (such as Claude Code, Codex, Antigravity, Aider, OpenCode) in AIM to prevent profile pollution, credential collisions, sidecar proxy failures, and broken lifecycle hooks.
---

# AIM Agent Adapter Onboarding Guide

## Overview

AIM (AI Multiplexer) enables developers to run multiple independent profiles (accounts, work contexts, client identities) for any AI coding agent on the same machine.

Every AI CLI agent operates under different assumptions regarding where it stores tokens, how it connects to language models, how it verifies lifecycle hooks, and how it handles system keychains. **An adapter's job is to guarantee complete profile isolation while preserving global developer convenience and ecosystem tooling.**

This guide codifies the architectural rules, lessons learned from Antigravity and Codex, and pre-flight verification gates required when implementing an adapter for any new agent (e.g. Claude Code, Aider, OpenCode, Hermes).

---

## When to Use

- Onboarding a brand-new AI CLI agent into AIM (e.g. implementing `internal/agents/<agent>/adapter.go`).
- Adding cross-profile conversation tracking (`internal/session/providers/<agent>/provider.go`).
- Debugging unexpected credential leakage, hook trust prompts, or connection errors when running an agent in an AIM profile.
- Auditing existing adapters for multi-profile safety and sidecar daemon resilience.

---

## The 8-Pillar Architecture for AI CLI Adapters

Every AI agent adapter in AIM MUST satisfy these 8 pillars:

```
  ┌────────────────────────────────────────────────────────────────────────┐
  │                    AIM Adapter Architecture Pillars                    │
  ├────────────────────────────────────┬───────────────────────────────────┤
  │ 1. Binary Discovery & Wrappers     │ 5. Hook Trust & Path Rewriting    │
  │ 2. Credential Isolation & Keychain │ 6. Transport & Sidecar Daemons    │
  │ 3. Config & XDG Dotfile Bridging   │ 7. Session & History Resumption   │
  │ 4. Plugins & Extension Ecosystem   │ 8. Self-Healing Diagnostics       │
  └────────────────────────────────────┴───────────────────────────────────┘
```

---

### Pillar 1: Binary Discovery & Wrapper Transparency

Agents may be installed globally via Homebrew (`/opt/homebrew/bin`), package managers (`~/.local/bin`, `npm -g`), or patched wrappers (e.g., `cxstatusline` wrapping `codex`).

**Rules:**
1. **Never hardcode binary paths**. Use `exec.LookPath(a.BinaryName())`.
2. **Support standard user fallbacks**: Check `~/.local/bin/<agent>` if `LookPath` fails in non-interactive shells.
3. **Respect wrapper scripts**: If a user has wrapped the binary (e.g., to inject environment variables or statuslines), execute the wrapper rather than bypassing it to find the raw upstream engine.

```go
func (a *Adapter) ResolveBinary() string {
    bin, err := exec.LookPath(a.BinaryName())
    if err == nil {
        return bin
    }
    fallback := filepath.Join(config.RealHomeDir(), ".local", "bin", a.BinaryName())
    if _, err := os.Stat(fallback); err == nil {
        return fallback
    }
    return a.BinaryName()
}
```

---

### Pillar 2: Credential Isolation & System Keychain Mitigation

Many CLI agents (Antigravity, Claude Code, GitHub Copilot) default to saving credentials in the macOS Keychain (`login.keychain-db`) using libraries like `keytar` or `go-keyring`.

**The Danger**: Because system keychains are global to the logged-in OS user, multiple profiles will overwrite each other's credentials in the host Keychain, causing cross-profile session corruption.

**Rules:**
1. **Identify the Storage Mode**: Does the agent support file-based token storage (like `auth.json` or `antigravity-oauth-token`)?
2. **Headless Bypass**: If the agent detects SSH or headless mode to bypass keychains, leverage it (e.g. `SSH_CONNECTION="127.0.0.1 50000 127.0.0.1 22"` used by Antigravity).
3. **Keychain Purge Registration**: If the agent writes to Keychain during initial OAuth login, register the service names in `internal/profile/keychain.go` and `internal/runner/exec.go`:
   - Auto-purge before launching unauthenticated profiles to prevent token inheritance.
   - Run a background harvest watcher to capture the token into the profile directory and immediately purge it from the host Keychain.

---

### Pillar 3: Configuration & XDG Directory Bridging

Agents often separate their state into:
- **Agent Home**: `~/.<agent>` or `~/.config/<agent>` (e.g., `~/.codex`, `~/.claude`).
- **Companion Tools**: External companion utilities, themes, or statuslines (e.g., `~/.config/cxstatusline`).

**Rules:**
1. **Isolate Agent State**: Redirect `HOME` and any agent-specific home variable (`CODEX_HOME`, `CLAUDE_HOME`) to `profileDir`.
2. **Bridge Companion Configs**: If the agent relies on an external tool configured under XDG directories (`~/.config/<tool>`), add that tool's path to `defaultBridgedPaths` in `internal/profile/symlink.go`.
3. **Non-Destructive Symlinking**: Never clobber existing profile files. If an external tool previously auto-generated a default directory (like fallback `settings.json`), detect and migrate it cleanly.

---

### Pillar 4: Plugins, Extensions & Skills Bridging

Developers expect their installed plugins, MCP servers, and custom skills to be available across all profiles without manual reinstallation.

**Rules:**
1. **Bridge Plugin Directories**: Symlink the host plugin directory into the profile:
   - `~/.codex/plugins` $\rightarrow$ `$profileDir/.codex/plugins`
   - `~/.gemini/config/plugins` $\rightarrow$ `$profileDir/.gemini/config/plugins`
   - `~/.claude/plugins` $\rightarrow$ `$profileDir/.claude/plugins`
2. **Bridge MCP Server Configurations**: Ensure `mcp_config.json` or `config.toml` MCP blocks are available.
3. **Shared Skills**: Always bridge `~/.agents/skills` via `symlink.go`.

---

### Pillar 5: Lifecycle Hooks & The Absolute-Path Trust Trap

Modern agents feature lifecycle hooks (`PreToolUse`, `PostCompact`, `SessionStart`, etc.) defined in JSON/YAML. To prevent malicious commands, agents store cryptographic approval hashes in their configuration file (e.g., `config.toml`).

**The Critical Trap**:
Agents often key trust hashes by the **absolute host path** of the hooks file:
```toml
[hooks.state."/Users/nemesis/.codex/hooks.json:pre_tool_use:0:0"]
trusted_hash = "sha256:abc..."
```
When AIM redirects `HOME` to `~/.aim/profiles/<profile>`, the agent checks for:
`"/Users/nemesis/.aim/profiles/<profile>/.codex/hooks.json:pre_tool_use:0:0"`!
If the paths do not match, the agent:
- Re-prompts the user to approve every single hook on startup.
- In headless/automated runs, blocks hook execution and blocks diagnostic commands with errors like `"Direct access to diagnostic information is blocked"`.

**The Solution (Dynamic Trust Path Rewriting):**
When copying or syncing the host configuration into the profile, always rewrite host hook paths to the profile's hook paths:

```go
func copyHostConfig(realHome, profileCodexDir string) {
    hostConfig := filepath.Join(realHome, ".codex", "config.toml")
    destConfig := filepath.Join(profileCodexDir, "config.toml")
    hostHooksJSON := filepath.Join(realHome, ".codex", "hooks.json")
    destHooksJSON := filepath.Join(profileCodexDir, "hooks.json")

    if data, err := os.ReadFile(destConfig); err == nil {
        if strings.Contains(string(data), hostHooksJSON) {
            rewritten := strings.ReplaceAll(string(data), hostHooksJSON, destHooksJSON)
            _ = os.WriteFile(destConfig, []byte(rewritten), 0644)
        }
    }
}
```

---

### Pillar 6: Transport Layer & Sidecar Reverse Proxies

Many developer setups use local reverse proxies or compression sidecars (e.g., Caveman on port 8787, Ollama on port 11434, local LLM mock servers).

**The Trap**:
The agent's configuration points to a localhost endpoint (`http://127.0.0.1:8787/chatgpt`).
- When running outside AIM, a supervisor script may launch the daemon.
- When running via AIM, AIM invokes the agent binary directly. If the daemon has terminated (e.g. after a 30-minute idle timeout), all agent requests fail with `Connection refused`.

**The Solution (Sidecar Probing & Auto-Start):**
In `PrepareEnv`:
1. Check if the agent's configuration routes traffic to a local proxy.
2. Probe the TCP port with a short timeout (`250ms`).
3. If unreachable, locate the daemon binary on the host and launch it as a detached process (`Setpgid: true`).
4. Wait up to `1.5s` for the port to accept connections before handing control to the agent.

```go
func ensureSidecarDaemons(realHome, profileDir string) {
    // 1. Inspect config for localhost base_url
    if !strings.Contains(cfgStr, "127.0.0.1:8787") {
        return
    }
    // 2. Probe port
    conn, err := net.DialTimeout("tcp", "127.0.0.1:8787", 250*time.Millisecond)
    if err == nil {
        conn.Close()
        return
    }
    // 3. Auto-start detached daemon
    cmd := exec.Command(bin)
    cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
    _ = cmd.Start()
    // 4. Poll until listening
}
```

---

### Pillar 7: Session Discovery & Cross-Profile Resumption

To support `aim sessions` and `aim resume <agent> <profile>`, the adapter must implement a session provider in `internal/session/providers/<agent>/provider.go`.

**Rules:**
1. **Identify Transcript Formats**: JSONL rollouts, SQLite databases, or Markdown logs.
2. **Metadata Extraction**: Extract session ID, timestamp, working directory, git branch, message count, and summary.
3. **Catalyst Context Handoff**: If resuming across profiles with different accounts, generate a Catalyst structured brief to carry over active task state.

---

### Pillar 8: Self-Healing Diagnostics (`Doctor`)

Every adapter's `Doctor()` method must report clear diagnostic results and auto-heal known configuration drifts:
1. **Binary Check**: Installed version and location.
2. **Auth Check**: Credential existence, email, and expiration time.
3. **Storage Check**: Verify isolated storage directory permissions (0700).
4. **Sidecar Check**: Verify required local proxy daemons are listening.
5. **Hook Trust Check**: Automatically rewrite un-migrated host hook paths so users never face blocked diagnostics.

---

## Step-by-Step Blueprint: Onboarding a New Agent (e.g., Claude Code)

When adding support for Claude Code (`claude`):

### Step 1: Discover Agent Footprint
Investigate the agent's files on the host:
- Binary: `which claude` $\rightarrow$ `/Users/.../.local/bin/claude`
- Config & State: `~/.claude.json` (global config), `~/.claude/` (history, cache, sessions)
- Auth Storage: macOS Keychain service `Claude Code-credentials` or `claude`
- Plugins: `~/.claude/plugins/`

### Step 2: Implement Adapter in `internal/agents/claude/adapter.go`
- `Name()`: `"claude"`
- `DisplayName()`: `"Claude Code"`
- `BinaryName()`: `"claude"`
- `PrepareEnv()`:
  - Create isolated `~/.aim/profiles/<profile>/.claude`
  - Bridge `~/.claude/plugins` $\rightarrow$ `$profileDir/.claude/plugins`
  - Copy and rewrite paths in `~/.claude.json`
  - Set `HOME=profileDir`
- `Doctor()`:
  - Validate binary, token existence, storage permissions, and keychain isolation status.

### Step 3: Register in AIM Core
1. Register in `cmd/aim/root.go` / `main.go`: `reg.Register(claude.NewAdapter())`.
2. Add keychain services (`Claude Code-credentials`, `claude`) to `internal/profile/keychain.go`.
3. Implement `internal/session/providers/claude/provider.go` to parse Claude session transcripts.

---

## Verification & Quality Checklist

Before submitting a PR for a new agent adapter, verify:

- [ ] **Clean Compilation & Race Testing**: `go test -race ./...` passes.
- [ ] **Multi-Profile Concurrency**: Can run two profiles of the agent simultaneously without credential overwrites.
- [ ] **Keychain Isolation**: `aim doctor` detects and purges host Keychain leaks without affecting active profiles.
- [ ] **Sidecar Auto-Start**: If a local proxy is configured, verify the agent starts even if the proxy was down.
- [ ] **Hook Trust**: Verify hooks run without interactive "untrusted hook" prompts in fresh profiles.
- [ ] **XDG Tooling**: Verify companion tools (themes, statuslines) load host configurations.
- [ ] **Integration Smoke Test**: Run `./test/smoke_test.sh` and ensure all suites pass.
