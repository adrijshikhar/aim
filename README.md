# aim — AI Multiplexer

`aim` is a lightweight CLI and TUI for running multiple AI coding assistants (Antigravity, Gemini, Codex, Claude…) under fully isolated profiles. Each profile gets its own virtual home directory — separate config, credentials, history, and state — so you can have a `work` and a `personal` Antigravity account side-by-side without them ever interfering.

```
aim run agy work          # run Antigravity under the "work" profile
aim run agy personal      # run Antigravity under the "personal" profile
aim run gemini research   # run Gemini under the "research" profile
aim usage                 # check remaining quota across all accounts
aim ui                    # open the interactive TUI dashboard
```

---

## Features

| Feature | Details |
|---|---|
| **Profile isolation** | Full virtual-home sandbox per profile — separate `~`, config, tokens, history |
| **Agent-aware profiles** | Profiles are tagged per agent; `aim list agy` shows only Antigravity profiles |
| **Multi-agent support** | Antigravity CLI (`agy`), Gemini CLI (`gemini`) out of the box; easily extensible |
| **Usage & quota tracking** | Live capacity gauges, 5h & weekly limits, reset countdowns, and instant caching |
| **OAuth PKCE login** | `aim login agy work` opens your browser, completes OAuth, and saves the token |
| **Interactive TUI** | Bubble Tea dashboard with dynamic usage gauges, inspector, tabs, and one-key launch |
| **Shell completions** | Tab-complete agents, profiles, and subcommands in zsh, bash, and fish |
| **Doctor command** | Diagnoses missing binaries, tokens, and misconfigurations per agent |

---

## Installation

### 1. Quick Install (macOS & Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/adrijshikhar/aim/main/install.sh | sh
```

This automatically detects your platform (`darwin`/`linux`, `arm64`/`amd64`), downloads the verified binary from GitHub Releases, checks SHA256 sums, and installs shell autocompletions for zsh, bash, and fish.

### 2. Homebrew

```bash
brew install adrijshikhar/tap/aim
```

### 3. Pre-Compiled Binaries

Download standalone archives for macOS and Linux from [GitHub Releases](https://github.com/adrijshikhar/aim/releases/latest).

### 4. Build from Source

Requires Go 1.23+:

```bash
git clone https://github.com/adrijshikhar/aim
cd aim
make install
```

### Activate completions after first install

**zsh** — add to `~/.zshrc` (once, if not already present):
```zsh
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit && compinit
```

**bash** — add to `~/.bashrc` (once):
```bash
source ~/.local/share/bash-completion/completions/aim
```

**fish** — nothing to do; fish picks up `~/.config/fish/completions/` automatically.

> **Tip:** If you installed with a custom `ZSH_COMPLETION_DIR`, adjust the `fpath` line accordingly.

---

## Usage

### Profiles

A profile is a named, sandboxed home directory. You can create as many as you like per agent.

```bash
aim login agy work         # authenticate Antigravity for "work"
aim login agy personal     # authenticate Antigravity for "personal"
aim login gemini research  # authenticate Gemini for "research"
```

### Running an agent

```bash
aim run agy work           # run Antigravity as "work"
aim run gemini research    # run Gemini as "research"
```

Pass extra flags after `--`:
```bash
aim run agy work -- --resume
```

### Listing profiles

```bash
aim list           # all profiles, showing associated agents and quota badges
aim list agy       # only profiles tagged for Antigravity
aim list gemini    # only profiles tagged for Gemini
```

Example output:
```
=== Configured Profiles ===
  1. work [agy] (5h: 82% [2h 15m], wk: 90% [5d 14h]) (~/.aim/profiles/work)
  2. personal [agy] (wk: 75% [3d 8h]) (~/.aim/profiles/personal)
  3. research [gemini] [no credentials] (~/.aim/profiles/research)
  4. offline-bot [agy] [offline] (~/.aim/profiles/offline-bot)
```

Profiles display inline usage badges when cached quota data is available:
- **Operational profiles**: display multi-window or single-window sliding quota and reset countdowns, such as `(5h: 82% [2h 15m], wk: 90% [5d 14h])` or `(wk: 75% [3d 8h])`. When fully charged (100%), the reset countdown is omitted (e.g. `(5h: 100%, wk: 100%)`).
- **Non-operational profiles**: display status badges such as `[no credentials]` when unauthenticated, or `[offline]` when provider APIs or local daemons are unreachable.

### Quota & Usage

Check remaining request limits, sliding windows, and reset countdowns across all configured accounts:

```bash
aim usage                    # inspect all profiles across all agents
aim usage agy                # filter by agent (Antigravity)
aim usage agy work           # check a specific profile
aim usage -r                 # force live refresh from providers (bypasses 3m cache)
aim usage --json             # output structured JSON array for scripting or CI
```

Example tabular output:
```
AGENT   PROFILE   STATUS  5H LIMIT       5H RESET  WEEKLY LIMIT   WEEKLY RESET  CHECKED
agy     work      OK      [████████  ] 82% 1h 45m   [██████████] 100% —           just now
agy     personal  OK      [██████████] 100% —      [██████████] 100% —           2m ago
gemini  research  NO_AUTH —              —         —              —             just now
```

Flags:
- `-r`, `--refresh`: Flush cached usage reports and query provider APIs directly.
- `--json`: Format usage reports as a machine-readable JSON array.

### Diagnostics

```bash
aim doctor           # check all agents and profiles
aim doctor agy       # check only Antigravity profiles
```

### Remove a profile

```bash
aim remove agy work       # disassociate "work" from agy (removes dir if no agents remain)
aim remove work           # remove "work" from all agents and delete its directory
```

### Interactive TUI

```bash
aim          # or: aim ui
```

The TUI provides an interactive dashboard with real-time quota telemetry:
- **Usage Gauges**: Profile rows display colored capacity gauges (green `>30%`, yellow `10–30%`, red `<10%`) for immediate quota visibility.
- **Inspector Pane**: Selecting any profile displays detailed 5-hour and weekly quota breakdown bars, reset countdowns, and last-checked timestamps.
- **Keybindings**:
  - `Tab` / `Shift+Tab`: Filter profiles by agent tab
  - `↑` / `↓` (`j` / `k`): Navigate profiles
  - `Enter`: Launch selected profile
  - `r`: Force live refresh of quota and usage metrics from providers
  - `q` / `Esc`: Quit

### Shell subshell

```bash
aim shell agy work    # drop into a subshell with agy's "work" environment
```

---

## Shell Auto-Completion

After `make install`, Tab completion is available for all subcommands:

```
aim run <Tab>           → agy  gemini
aim run agy <Tab>       → bot  personal  work
aim usage <Tab>         → agy  gemini
aim usage agy <Tab>     → bot  personal  work
aim login <Tab>         → agy  gemini
aim doctor <Tab>        → agy  gemini
aim remove <Tab>        → agy  gemini
aim completion <Tab>    → bash  fish  zsh
```

To regenerate and write completion scripts manually:

```bash
aim completion zsh   > ~/.zsh/completions/_aim
aim completion bash  > ~/.local/share/bash-completion/completions/aim
aim completion fish  > ~/.config/fish/completions/aim.fish
```

Or use the standalone Make target:
```bash
make install-completions
```

---

## Configuration

`aim` stores everything in `~/.aim/` (override with `AIM_HOME`):

```
~/.aim/
├── config.json          # profile registry and agent associations
└── profiles/
    ├── work/            # full virtual home for "work"
    │   └── .aim-agent   # agent association marker
    └── personal/        # full virtual home for "personal"
```

`config.json` example:
```json
{
  "profiles": {
    "work":     { "agents": ["agy"] },
    "personal": { "agents": ["agy"] },
    "research": { "agents": ["gemini"] }
  }
}
```

---

## Adding a New Agent

1. Create `internal/agents/<name>/adapter.go` implementing the `AgentAdapter` interface:
   ```go
   type AgentAdapter interface {
       Name() string
       DisplayName() string
       Aliases() []string
       BinaryName() string
       HasCredentials(profileDir string) bool
       Login(ctx context.Context, profileName, profileDir string) error
       PrepareEnv(profileName, profileDir string) (LaunchEnv, error)
       Doctor(ctx context.Context, profileName, profileDir string) []DiagnosticResult
       GetUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error)
   }
   ```
2. Register it in `cmd/aim/main.go`:
   ```go
   reg.Register(myagent.NewAdapter())
   ```

Completions, TUI tabs, and CLI filtering all pick up the new agent automatically.

---

## Development

```bash
make build      # compile binary to ./aim
make test       # run unit tests (with -race)
make smoke      # build + run integration smoke tests
make clean      # remove ./aim binary
```

---

## Uninstall

```bash
make uninstall  # removes binary and all shell completions
```

---

## Roadmap

See [ROADMAP.md](ROADMAP.md) for upcoming architectural milestones, including:
* **Headless Sub-Agent Bursting (`aim exec`):** Spawning background agents under alternate profiles to pool quotas.
* **Catalyst Integration:** Cross-agent typed handoff briefs (`.catalyst/handoff.json`) and agent skills.
* **Smart Quota Routing (`aim route`):** Dynamic load balancing based on live account limits.
* **Sessions & Cross-Profile Resumption (`aim sessions`, `aim resume`).**

---

## License

MIT

