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
| **Git & Forge CLIs** | `.gitconfig`, `.git-credentials`, `.config/git`, `.config/gh`, `.config/glab` | **Bridged** (symlinked from host) |
| **Security & Keys** | `.ssh`, `.gnupg`, `.netrc`, `Library/Keychains` (macOS) | **Bridged** (read/write access to host keys) |
| **Shell Configs** | `.zshrc`, `.bashrc`, `.profile`, `.config/fish` | **Bridged** (developer aliases & prompt settings) |
| **Package Managers** | `.npmrc`, `.yarnrc`, `.pip/pip.conf`, `.cargo/` | **Bridged** (package registry auth & configs) |
| **Cloud & Containers** | `.docker`, `.aws`, `.config/gcloud`, `.kube` | **Bridged** (cloud credentials) |
| **AIM State** | `.aim` | **Isolated** (blocked from profile symlinks) |
| **Antigravity State** | `.gemini/antigravity-cli/antigravity-oauth-token` | **Isolated** (per-profile token sandbox) |
| **Claude State** | `.claude`, `.claude.json` | **Isolated** (per-profile token sandbox) |
| **Codex State** | `.codex` | **Isolated** (per-profile token sandbox) |

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
  "default_agent": "agy",
  "default_profile": "personal",
  "custom_bridged_paths": [
    ".config/custom-tool",
    ".my-creds"
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

- **`custom_bridged_paths`**: Additional dotfile or config paths to bridge from the host home into every profile sandbox.
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
- **Dotfile Health**: Verifies that symlinks (such as `.gitconfig` and developer configs) resolve cleanly.
