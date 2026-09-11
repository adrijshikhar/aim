```text
    _    ___ __  __ 
   / \  |_ _|  \/  |
  / _ \  | || |\/| |   https://github.com/adrijshikhar/aim
 / ___ \ | || |  | |   AIM — AI Multiplexer
/_/   \_\___|_|  |_|
```

# aim — AI Multiplexer

`aim` is a lightweight CLI and TUI for multiplexing multiple AI coding assistants (Antigravity, Gemini, Claude Code, Codex…) under fully isolated profiles with real-time quota telemetry.

Each profile gets its own sandboxed virtual home directory — separate credentials, history, and state — so you can switch between `work` and `personal` accounts seamlessly.

```bash
aim                      # open the interactive TUI dashboard
aim run agy work         # run Antigravity under the "work" profile
aim run agy personal     # run Antigravity under the "personal" profile
aim usage                # check quota capacity and reset timers across accounts
aim doctor               # diagnose binaries, tokens, and dotfile health
```

---

## Features

| Feature | Details |
|---|---|
| **Profile isolation** | Full virtual-home sandbox per profile — separate `~`, config, tokens, history |
| **Agent-aware profiles** | Profiles are tagged per agent; `aim list agy` shows only Antigravity profiles |
| **Multi-agent support** | Antigravity CLI (`agy`), Gemini CLI (`gemini`) out of the box; easily extensible |
| **Usage & quota tracking** | Live capacity gauges, 5h & weekly limits, reset countdowns, and instant caching |
| **OAuth PKCE login** | `aim login agy work` completes OAuth in browser and isolates the token |
| **Interactive TUI** | Bubble Tea dashboard with real-time capacity gauges, tabs, and one-key launch |
| **Shell completions** | Tab-complete agents, profiles, and subcommands in zsh, bash, and fish |
| **Doctor command** | Diagnoses missing binaries, tokens, and misconfigurations per agent |

---

## Installation

### Homebrew (macOS & Linux)

```bash
brew install adrijshikhar/tap/aim
```

### Quick Install (Script)

```bash
curl -fsSL https://raw.githubusercontent.com/adrijshikhar/aim/main/install.sh | sh
```

### Pre-Compiled Binaries & Source
Download standalone archives from [GitHub Releases](https://github.com/adrijshikhar/aim/releases/latest), or build from source with Go 1.23+:
```bash
git clone https://github.com/adrijshikhar/aim.git && cd aim && make install
```

---

## CLI Cheatsheet

| Command | Description |
|---|---|
| `aim` / `aim ui` | Open the interactive TUI dashboard |
| `aim run <agent> <profile> [-- args...]` | Execute an agent under an isolated profile |
| `aim usage [agent] [profile] [-r] [--json]` | Display remaining quota, reset timers, and credits |
| `aim list [agent]` | List all configured profiles and inline quota badges |
| `aim login <agent> <profile>` | Authenticate a new account via OAuth PKCE |
| `aim whoami` | Show active profile, agent, session ID, and quota info |
| `aim doctor [agent]` | Check environment, binary paths, tokens, and ADC status |
| `aim shell <agent> <profile>` | Launch an isolated subshell with profile environment |
| `aim clone <agent> <src> <dst>` | Duplicate profile settings without copying tokens |
| `aim remove [agent] <profile>` | Unlink agent from profile (deletes dir if empty) |
| `aim completion <shell>` | Generate shell completions (`zsh`, `bash`, `fish`) |

---

## Documentation

- 📖 **[Wiki & Architecture Guide](wiki.md)** — In-depth guide to virtual home isolation, dotfile bridging rules, directory structures, and configuration schemas.
- 🛠️ **[Contributing Guide](CONTRIBUTING.md)** — Local development setup, test suite execution (`make test`, `make smoke`), and how to add new agent adapters.
- 🗺️ **[Roadmap](ROADMAP.md)** — Upcoming milestones including quota pooling, headless bursting, and cross-agent resumption.

---

## License

[MIT](LICENSE)
