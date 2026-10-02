# aim — AI Multiplexer

`aim` is a lightweight CLI and TUI for running multiple AI coding assistants (Antigravity, Gemini, Claude Code, Codex) under fully isolated profiles with real-time quota telemetry.

Each profile gets its own sandboxed virtual home directory — separate credentials, history, and state — so you can switch between `work` and `personal` accounts seamlessly.

<p align="center">
  <img src="assets/tui-dashboard.png" alt="aim TUI Dashboard" width="850" />
</p>

```bash
aim                      # open the interactive TUI dashboard
aim run agy work         # run Antigravity under the "work" profile
aim run codex personal   # run Codex under the "personal" profile
aim usage agy            # check quota capacity and reset timers across accounts
aim doctor               # diagnose binaries, tokens, and dotfile health
```

---

## Features

- **Profile isolation** — a virtual `~` per profile with its own config, tokens, and history; dotfiles like `.gitconfig` and `.ssh` are bridged from the host ([details](wiki.md#what-is-isolated-vs-bridged)).
- **Four agents** — Antigravity (`agy`), Gemini (`gemini`), Claude Code (`claude`), and Codex (`codex`).
- **Interactive TUI** — profile list, agent tabs, fuzzy filter, live quota gauges, one-key launch.
- **Quota telemetry** — 5h and weekly limits, reset countdowns, and credits across every account.
- **Sessions & handoffs** — browse history across profiles, resume verbatim, or hand off context with [Catalyst](https://github.com/adrijshikhar/catalyst).
- **Shared host MCP servers & plugins** — your normal agent config is available in every profile.

---

## Installation

**Homebrew** (macOS & Linux):
```bash
brew install --cask adrijshikhar/tap/aim     # upgrade: brew upgrade adrijshikhar/tap/aim
```

**Script** (pin a release with `AIM_VERSION=v0.10.0`):
```bash
curl -fsSL https://raw.githubusercontent.com/adrijshikhar/aim/main/install.sh | sh
```

**From source** (Go 1.27.1+), or grab an archive from [GitHub Releases](https://github.com/adrijshikhar/aim/releases/latest):
```bash
git clone https://github.com/adrijshikhar/aim.git && cd aim && make install
```

> [!TIP]
> Homebrew clears the macOS quarantine flag for you. For browser-downloaded archives, run `xattr -d com.apple.quarantine $(which aim)`.

---

## Usage

| Command | Description |
|---|---|
| `aim` | Open the interactive TUI dashboard |
| `aim run <agent> <profile> [-- args...]` | Run an agent under an isolated profile |
| `aim login <agent> <profile>` | Authenticate an account via OAuth PKCE |
| `aim list [agent]` | List profiles with inline quota badges |
| `aim usage [agent] [profile] [-r] [--json]` | Remaining quota, reset timers, and credits |
| `aim sessions [agent] [--active] [--json]` | Active and past sessions across profiles and host |
| `aim resume <agent> <profile> [id] [-- args...]` | Resume verbatim (`--exact`) or via Catalyst (`--catalyst`) |
| `aim whoami` | Active profile, agent, session, and quota |
| `aim doctor [agent]` | Check binaries, tokens, ADC status, and keychain isolation |
| `aim mv` | Move an agent account and credentials to another profile |
| `aim remove [agent] <profile>` | Delete profile credentials and state |
| `aim completion <shell>` | Shell completions (`zsh`, `bash`, `fish`) |
| `aim version` | Show the aim version |

Add `--debug` to any command for verbose tracing.

**`aim sessions`** — active and recent conversations across profiles and host:

<p align="center">
  <img src="assets/cli-sessions.png" alt="aim sessions CLI Output" width="850" />
</p>

**`aim usage`** — quota limits, capacity gauges, and reset countdowns across accounts:

<p align="center">
  <img src="assets/cli-usage.png" alt="aim usage Telemetry Table" width="850" />
</p>

### TUI

Press `?` in the dashboard for every shortcut. The essentials: `↑`/`↓` select, `Enter` launch, `Tab` switch agent, `/` filter, `s` sessions, `q` quit. Full list in the [wiki](wiki.md#4-interactive-tui-dashboard).

---

## Host servers and plugins

MCP servers and enabled plugins from your normal agent config (the host) are merged into a profile while its sessions run, and removed when the last one exits.

| Agent | MCP servers | Plugins |
|---|---|---|
| Claude | `mcpServers` in `~/.claude.json` | `enabledPlugins`, `extraKnownMarketplaces` in `~/.claude/settings.json` |
| Codex | `[mcp_servers.*]` in `~/.codex/config.toml` | `[plugins."<id>"]` in the same file |
| agy | `mcpServers` in the shared `mcp_config.json` | — (already shared) |

If a session changes a server or plugin, aim asks at exit whether to promote it to the host or keep it in the profile. Turn the merge off per profile in `~/.aim/config.json` with `"mcp_global": false` or `"plugins_global": false`.

Prompts, overrides, backups, and edge cases: see the [wiki](wiki.md#8-host-servers--plugins).

---

## Session Resumption & Catalyst Handoffs

`aim sessions` (or `s` in the TUI) lists sessions across profiles and the host.

<p align="center">
  <img src="assets/tui-sessions.png" alt="aim Sessions Explorer Drawer" width="850" />
</p>

- **Exact resume** (`Enter` / `--exact`) — reopens the session with its full history.
- **Catalyst handoff** (`c` / `--catalyst`) — distills goal, decisions, and notes into a brief (`.catalyst/handoffs/<branch>.json`) so another profile or agent continues with a fresh context window.

Thanks to **[Catalyst](https://github.com/adrijshikhar/catalyst)** for the handoff protocol.

---

## Documentation

- 📖 **[Wiki](wiki.md)** — isolation model, directory layout, config schema, TUI, host merge.
- 🩺 **[Troubleshooting](TROUBLESHOOTING.md)** — OAuth, session resumption, SQLite, macOS signing.
- 🎨 **[Design System](design.md)** — palette, tokens, Lipgloss standards.
- 🛠️ **[Contributing](CONTRIBUTING.md)** — dev setup, tests, builds, new agent adapters.

---

## License

[MIT](LICENSE)
