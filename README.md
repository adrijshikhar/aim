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
| **Multi-agent support** | Antigravity CLI (`agy`), Gemini CLI (`gemini`), and OpenAI Codex (`codex`) out of the box; easily extensible |
| **Keychain isolation** | System keychains mounted for developer tools (`gh`, `git`); agent tokens explicitly purged |
| **Debug logging** | Opt-in tracing via `--debug`, `AIM_DEBUG=1`, or config file with dual console/file logs |
| **Usage & quota tracking** | Live capacity gauges, 5h & weekly limits, reset countdowns, and instant caching |
| **OAuth PKCE login** | `aim login agy work` completes OAuth in browser and isolates the token |
| **Interactive TUI** | Bubble Tea dashboard with real-time capacity gauges, tabs, and one-key launch |
| **Shell completions** | Tab-complete agents, profiles, and subcommands in zsh, bash, and fish |
| **Doctor command** | Diagnoses missing binaries, tokens, ADC status, and keychain isolation |

---

## Installation

### Homebrew (Recommended for macOS & Linux)

The easiest way to install and stay up to date:

```bash
brew tap adrijshikhar/homebrew-tap
brew install --cask aim
```

*Or via single command:*
```bash
brew install --cask adrijshikhar/tap/aim
```

To update to future releases:
```bash
brew update && brew upgrade adrijshikhar/tap/aim
```

> [!TIP]
> **macOS Gatekeeper Safe**: Installing via Homebrew automatically manages permissions and clears the quarantine flag (`com.apple.quarantine`), so `aim` runs without "unidentified developer" security warnings.

### Quick Install (Script)

```bash
curl -fsSL https://raw.githubusercontent.com/adrijshikhar/aim/main/install.sh | sh
```

### Pre-Compiled Binaries & Source
Download standalone archives from [GitHub Releases](https://github.com/adrijshikhar/aim/releases/latest), or build from source with Go 1.23+:
```bash
git clone https://github.com/adrijshikhar/aim.git && cd aim && make install
```

*(If downloading standalone archives directly in a macOS web browser, run `xattr -d com.apple.quarantine $(which aim)` to clear Gatekeeper warnings).*

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
| `aim doctor [agent]` | Check environment, binary paths, tokens, ADC status, and keychain isolation |
| `aim shell <agent> <profile>` | Launch an isolated subshell with profile environment |
| `aim clone <agent> <src> <dst>` | Duplicate profile settings without copying tokens |
| `aim remove [agent] <profile>` | Unlink agent from profile (deletes dir if empty) |
| `aim completion <shell>` | Generate shell completions (`zsh`, `bash`, `fish`) |
| `aim --debug <command>` | Run any command with verbose debug tracing |

---

## Documentation

- 📖 **[Wiki & Architecture Guide](wiki.md)** — In-depth guide to virtual home isolation, dotfile bridging rules, directory structures, and configuration schemas.
- 🛠️ **[Contributing Guide](CONTRIBUTING.md)** — Local development setup, test suite execution (`make test`, `make smoke`), and how to add new agent adapters.
- 🗺️ **[Roadmap](ROADMAP.md)** — Upcoming milestones including quota pooling, headless bursting, and cross-agent resumption.

---

## License

[MIT](LICENSE)
