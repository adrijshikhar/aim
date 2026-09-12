# AIM (AI Multiplexer) — Wiki & Architectural Guide

Welcome to the AIM Wiki. This guide covers detailed architecture, directory structures, configuration specifications, and advanced operational workflows.

---

## 1. Profile Isolation & Virtual Home Architecture

AIM provides isolated sandboxes for AI coding assistants by virtualizing the user's `HOME` directory. When an agent is executed under a profile (e.g. `aim run agy work`), AIM redirects `$HOME` to:

```text
~/.aim/profiles/<profile>/
```

### What is Isolated vs. Bridged

To balance complete agent isolation with seamless developer workflow, AIM selectively bridges global developer tools while strictly walling off agent credentials:

| Category | Path | Isolation Behavior |
|---|---|---|
| **Security & Keys** | `.ssh`, `.gnupg`, `.netrc` | **Bridged** (read/write access to host keys) |
| **System Keychains** | `Library/Keychains` (macOS) | **Bridged with Agent Ignore List** (mounted for `gh`, `git`, certs; agent services purged) |
| **Shell Configs** | `.zshrc`, `.bashrc`, `.profile`, `.config/fish` | **Bridged** (developer aliases & prompt settings) |
| **Package Managers** | `.npmrc`, `.yarnrc`, `.pip/pip.conf`, `.cargo/` | **Bridged** (package registry auth & configs) |
| **Cloud & Containers** | `.docker`, `.aws`, `.config/gcloud`, `.kube` | **Bridged** (cloud credentials) |
| **AIM State** | `.aim` | **Isolated** (blocked from profile symlinks) |
| **Antigravity State** | `.gemini/antigravity-cli/antigravity-oauth-token` | **Isolated** (per-profile token sandbox) |
| **Claude State** | `.claude`, `.claude.json` | **Isolated** (per-profile token sandbox) |
| **Codex State** | `.codex` | **Isolated** (per-profile token sandbox) |

### macOS Keychain Architecture & Agent Ignore List

On macOS, AIM mounts (symlinks) `Library/Keychains` into each profile sandbox so developer CLI utilities (such as `gh`, `git-credential-osxkeychain`, SSL/TLS system certificates, and package manager credentials) continue to function seamlessly.

However, CLI agents such as Antigravity (`agy`), Claude Code (`claude`), and OpenAI Codex (`codex`) store credentials in the macOS Keychain (`login.keychain-db`) using libraries like `go-keyring` or `node-keytar`. Left unmanaged, multiple profiles would overwrite each other's credentials in the host login keychain.

To solve this permanently without breaking developer tools, AIM enforces an **Explicit Agent Ignore List**:
- **Automatic Scrubbing**: Before launching an agent, during profile initialization, on login, and during `aim doctor`, AIM scrubs entries matching known agent services from the Keychain:
  - **Antigravity / Gemini**: `gemini` (account `antigravity`), `antigravity`, `antigravity-oauth-token`
  - **Claude Code**: `claude`, `claude-code`, `@anthropic-ai/claude-code`, `Claude Code-credentials`
  - **OpenAI Codex**: `codex`, `openai-codex`, `openai`
- **Forced Token File Sandboxing**: Purging these shared keychain entries forces the agent to read and write tokens strictly within its profile directory (`~/.aim/profiles/<profile>/...`).
- **Post-Execution Cleanup**: AIM runs deferred cleanup when an agent exits to ensure no session credentials linger in the macOS Keychain.
- **Custom Ignore List**: Users can add custom keychain service names via `"custom_ignored_keychains"` in `~/.aim/config.json`.

### Antigravity Shared State vs. Token Isolation

For Antigravity CLI (`agy`), AIM maintains a unified session cache while isolating credentials:
- **Shared across profiles**: Conversations, brainstorm trajectories (`brain/`), conversation database (`conversation_summaries.db`), and command history (`history.jsonl`) are automatically bridged to `~/.gemini/antigravity-cli/` or `~/.aim/shared/antigravity-cli/`.
- **Isolated per profile**: `antigravity-oauth-token` is stored strictly inside `~/.aim/profiles/<profile>/.gemini/antigravity-cli/antigravity-oauth-token`.

---

## 2. Directory Layout & Configuration

Everything managed by AIM resides in `~/.aim/` (configurable via the `AIM_HOME` environment variable):

```text
~/.aim/
├── config.json                 # Central profile registry & agent associations
├── aim-debug.log               # Persistent debug trace log (when debug is enabled)
├── cache/
│   └── usage_cache.json        # 3-minute TTL quota and rate-limit cache
├── shared/
│   └── antigravity-cli/        # Shared session history & conversation DB
└── profiles/
    ├── work/                   # Virtual home for "work" profile
    │   ├── .gitconfig -> ...   # Symlinked developer configs
    │   ├── .ssh -> ...
    │   └── .gemini/
    │       └── antigravity-cli/
    │           └── antigravity-oauth-token # Isolated work token
    └── personal/               # Virtual home for "personal" profile
        └── .gemini/
            └── antigravity-cli/
                └── antigravity-oauth-token # Isolated personal token
```

### `config.json` Specification

```json
{
  "debug": false,
  "default_agent": "agy",
  "default_profile": "personal",
  "custom_bridged_paths": [
    ".config/custom-tool",
    ".my-creds"
  ],
  "custom_ignored_keychains": [
    "custom-agent-auth"
  ],
  "profiles": {
    "work": {
      "agents": ["agy", "claude"],
      "env": {
        "ENV_OVERRIDE": "production"
      },
      "args": ["--dangerously-skip-permissions"]
    },
    "personal": {
      "agents": ["agy"]
    }
  }
}
```

- **`debug`**: Enable verbose debug logging to stderr and `~/.aim/aim-debug.log` (`true` / `false`, default: `false`). Can also be toggled via `AIM_DEBUG=1` or `--debug`.
- **`custom_bridged_paths`**: Additional dotfile or config paths to bridge from the host home into every profile sandbox.
- **`custom_ignored_keychains`**: List of additional macOS Keychain service names to scrub before and after agent execution to maintain strict profile isolation.
- **`env`**: Profile-specific environment variables injected on launch.
- **`args`**: Extra CLI arguments automatically passed to the agent binary when launched under this profile.

---

## 3. Shell Completion Setup

Shell completions provide dynamic Tab completion for commands, agents, and configured profiles.

### Automatic Activation

If installed via Homebrew or the curl installer, completions are placed in standard directories:
- **zsh**: `~/.zsh/completions/_aim`
- **bash**: `~/.local/share/bash-completion/completions/aim`
- **fish**: `~/.config/fish/completions/aim.fish`

### Manual Shell Configuration

**zsh** — Ensure `fpath` includes the completion directory in `~/.zshrc`:
```zsh
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit && compinit
```

**bash** — Source the completion file in `~/.bashrc`:
```bash
source ~/.local/share/bash-completion/completions/aim
```

**fish** — Automatically discovered from `~/.config/fish/completions/aim.fish`.

### Manual Script Generation

```bash
aim completion zsh  > ~/.zsh/completions/_aim
aim completion bash > ~/.local/share/bash-completion/completions/aim
aim completion fish > ~/.config/fish/completions/aim.fish
```

---

## 4. Interactive TUI Dashboard

Launching `aim` without arguments (or running `aim ui`) opens the terminal user interface built with Charm Bubble Tea & Lip Gloss.

### TUI Features
- **Profile Navigation**: Use `↑` / `↓` (`k` / `j`) to browse configured profiles.
- **Agent Tabs**: Use `Tab` / `Shift+Tab` or numbers `[1]`, `[2]`... to filter profiles by agent.
- **Live Quota Gauges**: Color-coded capacity indicators:
  - 🟢 **Green (`>30%`)**: Ample quota available.
  - 🟡 **Yellow (`10%–30%`)**: Approaching threshold.
  - 🔴 **Red (`<10%`)**: Depleted or near rate limit.
- **Inspector Drawer**: Selecting a profile displays 5-hour limit, weekly limit, reset countdowns, credits remaining, and cache freshness.
- **Actions**:
  - `[Enter]`: Launch selected profile immediately.
  - `[s]`: Drop into an isolated subshell with the profile environment.
  - `[l]`: Launch browser OAuth login to authenticate the profile.
  - `[d]`: Open the embedded Doctor diagnostics drawer.
  - `[m]` / `[R]`: Open the interactive profile rename modal.
  - `[x]`: Open the interactive profile deletion modal.
  - `[r]`: Bypass cache and force a live quota refresh.
  - `[q]`: Quit.

---

## 5. Doctor Diagnostics

The `aim doctor` command runs comprehensive pre-flight diagnostics:
- **Binary Discovery**: Verifies that required agent binaries (`agy`, `gemini`, `claude`) exist in `PATH` or `~/.local/bin`.
- **Token Integrity**: Checks token file existence, permissions (recommended `0600`), and JSON structure validity.
- **Token Expiry**: Validates access token expiration and refresh token availability for auto-renewal.
- **ADC Detection**: Confirms Google Cloud Application Default Credentials if active.
- **Keychain Isolation (macOS)**: Scans the macOS Keychain for lingering agent credentials and automatically purges them to guarantee clean profile isolation.
- **Dotfile Health**: Verifies that symlinks (such as `.gitconfig` and developer configs) resolve cleanly.

---

## 6. Debug Mode & Operational Tracing

AIM includes a verbose debug logging subsystem for diagnosing environment setup, process spawning, token discovery, and dotfile bridging.

### Opt-In Mechanisms

Debug logging is **off by default**. You can enable it via any of the following methods:

1. **CLI Flag**: Pass `--debug` to any `aim` command:
   ```bash
   aim --debug run agy work
   aim --debug whoami
   aim --debug doctor
   ```

2. **Environment Variable**: Set `AIM_DEBUG=1` (or `true`, `on`, `yes`):
   ```bash
   export AIM_DEBUG=1
   aim list
   ```

3. **Global Config**: Set `"debug": true` in `~/.aim/config.json`:
   ```json
   {
     "default_agent": "agy",
     "default_profile": "work",
     "debug": true
   }
   ```

### Output Channels

When debug mode is active:
- **Console Output**: Color-coded, timestamped trace messages are streamed directly to `os.Stderr` (automatically suppressed during interactive TUI sessions to prevent screen corruption).
- **Persistent Log File**: All debug traces are appended to `~/.aim/aim-debug.log` with microsecond timestamps and severity levels.
