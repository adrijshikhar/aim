# AIM (AI Multiplexer) — Wiki & Architectural Guide

Welcome to the AIM Wiki. This guide covers detailed architecture, directory structures, configuration specifications, and advanced operational workflows.

---

## 1. Profile Isolation & Virtual Home Architecture

AIM provides isolated sandboxes for AI coding assistants by virtualizing the user's `HOME` directory. When an agent is executed under a profile (e.g. `aim run agy work`), AIM redirects `$HOME` to:

```text
~/.aim/profiles/<profile>/
```

### What is Isolated vs. Bridged

To balance complete agent isolation with seamless developer workflow, AIM shares the host home with every profile and walls off only agent state. On every launch (and on `aim doctor`) it links **every top-level dot entry of your real home** into the profile, except a fixed deny-list. A tool you install on the host after a profile was created is picked up on its next launch; there is no allow-list to extend.

| Category | Path | Isolation Behavior |
|---|---|---|
| **Host dotfiles** | every `~/.*` not denied below: `.gitconfig`, `.ssh`, `.gnupg`, `.config`, `.local`, `.cargo`, `.npmrc`, `.docker`, `.aws`, `.kube`, `.agents`, `.mcp-auth`, … | **Bridged** (one link per top-level entry; a dir that contains the profiles, such as `.local` on the default XDG layout where profiles live in `~/.local/share/aim/profiles`, is not linked itself: its children are, e.g. `.local/bin`, `.local/state`, `.local/share/claude`, down to but excluding the aim data dir) |
| **Go path** | `go` (only if `~/go` exists) | **Bridged** (module and build cache shared, not duplicated per profile) |
| **System Keychains** | `Library/Keychains` (macOS) | **Bridged with Agent Ignore List** (mounted for `gh`, `git`, certs; agent services purged) |
| **User caches** | `Library/Caches` (macOS) | **Bridged** (go-build, pip, Homebrew, Playwright, bun, … shared) |
| **Claude extensions** | `.claude/plugins`, `.claude/skills`, `.claude/rules`, `.claude/commands`, `.claude/hooks` | **Bridged** inside the per-profile `.claude` |
| **Per-app state** | `Library` itself, `Library/Application Support`, `Library/Preferences` | **Per profile** (not bridged) |
| **AIM State** | `.aim` | **Isolated** (never linked) |
| **Antigravity / Gemini State** | `.gemini` | **Isolated** (per-profile token sandbox; agy shares its conversation history separately, see below) |
| **Claude State** | `.claude`, `.claude.json`, `.claude.json.*` (backups, temp files) | **Isolated** (per-profile token sandbox) |
| **Codex State** | `.codex` | **Isolated** (per-profile token sandbox) |
| **Host-only noise** | `.Trash`, `.DS_Store`, `.CFUserTextEncoding`, `.localized`, `*_history` (`.zsh_history`, `.python_history`, …), `.zsh_sessions`, `.bash_sessions`, `.zcompdump*` | **Not bridged** (each profile keeps its own) |

`.mcp-auth` (OAuth tokens that `mcp-remote` caches for remote MCP servers) is shared, so an MCP server authorised on the host works in every profile. Sockets and fifos in the home are never linked.

**Overriding per profile.** A real, non-empty file or dir in the profile at a bridged path is a profile override and is always kept: to give one profile its own `~/.npmrc` or `~/.config`, put it there. Only empty or known stub dirs (and dirs that hold nothing but links into the matching host dir, as left by older AIM versions) are replaced with a link; nothing else is ever removed. When a profile overrides a parent such as `.config`, `.cargo`, `.local` or `.pip`, AIM still links the credential-bearing children it knows about (`.config/gh`, `.config/git`, `.config/glab`, `.config/gcloud`, `.config/fish`, `.cargo/config.toml`, `.cargo/credentials.toml`, `.pip/pip.conf`, …) inside it. AIM never creates or removes anything below a profile path that is itself a link to the host. `aim doctor` lists every profile copy that shadows a host path, with the `rm -rf` that restores sharing.

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

### OpenAI Codex CLI (`codex`) Architecture

For OpenAI Codex CLI (`codex`), AIM utilizes Codex's native environment virtualization:
- **`$CODEX_HOME` Redirection**: AIM sets `CODEX_HOME=~/.aim/profiles/<profile>/.codex` and `HOME=~/.aim/profiles/<profile>`.
- **Pure File-Based Auth**: Unlike tools reliant on OS Keychains, Codex reads and writes its authentication state strictly to `$CODEX_HOME/auth.json` (mode `0600`).
- **Concurrent Execution**: Because each profile possesses its own isolated SQLite databases (`state_5.sqlite`, `logs_2.sqlite`), session history, and config, multiple profiles can run concurrently in separate terminals without database lock collisions.
- **Account & Quota Telemetry**: AIM extracts the authenticated user email and ChatGPT plan type (e.g. `ChatGPT Plus`, `ChatGPT Pro`, `ChatGPT Team`) directly from the JWT `id_token` payload claims, and parses session logs for real-time rate limit telemetry.

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
    "Library/Application Support/custom-tool"
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
- **`custom_bridged_paths`**: Additional host paths to bridge into every profile sandbox. Top-level dotfiles are already bridged, so this is for paths outside that scan, such as `Library/Application Support/<tool>`. Paths that escape the home or name agent state (`.aim`, `.claude`, `.claude.json*`, `.codex`, `.gemini`) are refused.
  A non-empty profile copy of a bridged path is kept as an override, so the host's is silently not used; `aim doctor` flags each such copy (and any link pointing elsewhere) with the `rm -rf` that restores sharing — it never deletes anything itself.
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

Launching `aim` without arguments opens the terminal user interface built with Charm Bubble Tea & Lip Gloss.

### TUI Features
- **Profile Navigation**: Use `↑` / `↓` (`k` / `j`) to browse configured profiles.
- **Agent Tabs**: Use `Tab` / `Shift+Tab` or numbers `[1]`, `[2]`... to filter profiles by agent.
- **Live Quota Gauges**: Color-coded capacity indicators:
  - 🟢 **Green (`>30%`)**: Ample quota available.
  - 🟡 **Yellow (`10%–30%`)**: Approaching threshold.
  - 🔴 **Red (`<10%`)**: Depleted or near rate limit.
- **Inspector Drawer**: Selecting a profile displays 5-hour limit, weekly limit, reset countdowns, credits remaining, and cache freshness.
- **Keybindings** (press `?` in the dashboard for the same list):

| Key | Action |
|---|---|
| `↑` / `k`, `↓` / `j` | Navigate profile list |
| `Enter` | Launch selected profile |
| `Tab` / `Shift+Tab` | Cycle agent tabs |
| `1` – `3` | Jump to the first three agent tabs |
| `/` | Live fuzzy filter (profile or agent name) |
| `Esc` | Clear filter or close modal / drawer |
| `s` | Sessions Explorer drawer (resume exact or Catalyst handoff) |
| `l` | Browser OAuth login for the selected profile |
| `d` | Doctor diagnostics drawer |
| `F` | In-tool feedback modal (general, bug report, or feature request) |
| `m` / `R` | Rename selected profile |
| `v` | Move selected profile |
| `x` / `Delete` | Delete / unlink selected profile (with confirmation) |
| `r` | Bypass cache and force a live quota refresh |
| `?` | Toggle help overlay |
| `q` / `Ctrl+C` | Quit |

---

## 5. Doctor Diagnostics

The `aim doctor` command runs comprehensive pre-flight diagnostics:
- **Binary Discovery**: Verifies that required agent binaries (`agy`, `gemini`, `claude`) exist in `PATH` or `~/.local/bin`.
- **Token Integrity**: Checks token file existence, permissions (recommended `0600`), and JSON structure validity.
- **Token Expiry**: Validates access token expiration and refresh token availability for auto-renewal.
- **ADC Detection**: Confirms Google Cloud Application Default Credentials if active.
- **Keychain Isolation (macOS)**: Scans the macOS Keychain for lingering agent credentials and automatically purges them to guarantee clean profile isolation.
- **Dotfile Health**: Verifies that symlinks (such as `.gitconfig` and developer configs) resolve cleanly.
- **Sanitized Markdown Report (`aim doctor --report`)**: Generates an anonymized diagnostic report formatted in Markdown, scrubbing home directory paths (`~`) and redacting configured profile names (`[profile-1]`, `[profile-2]`), ready for GitHub issue reporting.

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

---

## 7. Sessions, Cross-Profile Resumption & Catalyst Handoff

AIM provides comprehensive conversation session discovery, cross-profile thread hydration, and vendor-neutral Catalyst handoffs across all configured profiles and host dotfiles.

### 7.1 Multi-Source Discovery & Process Correlation

AIM inspects conversation storage across both virtualized profile homes (`~/.aim/profiles/*/`) and unmanaged host environments (`~/.gemini/antigravity-cli`, `~/.codex`):
- **Antigravity (`agy`)**: Reads `conversation_summaries.db` extracting `conversation_id`, title, preview text, and last modified timestamps.
- **Codex (`codex`)**: Scans `$CODEX_HOME/state_5.sqlite` (`threads` table) and falls back to JSONL index records (`session_index.jsonl`) and session rollout files (`sessions/*.jsonl`).
- **Live Process Scanner**: Rather than relying on stale 0-byte `.lock` files, AIM inspects the active OS process table (`ps -eo pid,command`) matching `--conversation=<id>` (Antigravity) and `resume <id>` (Codex). Active processes are automatically marked with `Status: ACTIVE (PID <pid>)`.

### 7.2 Dual-Mode Resumption

When resuming a conversation under a destination profile:
1. **Exact Thread (`--exact` or `[Enter]` in TUI)**:
   - Resumes the verbatim conversation history and local state.
   - If resuming a host thread into a profile, AIM automatically hydrates the thread (`Hydrate`) into the destination profile's database and session storage.
   - Passing `--fork` (`-b`) generates a clean child conversation ID to branch off without mutating the original history.
2. **Catalyst Summary Handoff (`--catalyst`, `-c`, or `[c]` in TUI)**:
   - Extracts goal, decisions, and trajectory context into Catalyst's standard `.catalyst/handoffs/<branch>.json` brief.
   - Starts a fresh context window under the destination profile, primed with the condensed handoff brief. This eliminates context-rot and allows cross-vendor resumption (e.g. continuing an Antigravity task inside Codex).

### 7.3 Codex Hook & Plugin Bridging

To ensure Catalyst handoff hooks fire reliably inside isolated Codex profiles, AIM automatically bridges:
- Host `~/.codex/plugins/` → Profile `.codex/plugins/`
- Host `~/.codex/hooks.json` & `~/.codex/hooks/` → Profile `.codex/hooks/` and `.codex/hooks.json`

### 7.4 CLI Commands

- `aim sessions [agent]`: Lists active and recent conversations formatted as tables, with `--profile`, `--agent`, `--active`, `--all`, and `--json` options (launches the interactive TUI Sessions Explorer in a terminal).
- `aim resume <agent> <profile> [session-id]`: Resumes a session with prefix matching (e.g. `aim resume agy work 775e6ada`), active process collision warnings, and `--catalyst` or `--exact` modes.
- **Interactive TUI Drawer Preview**: In the TUI Sessions Drawer (`s`), an instant (0ms) preview box displays the full summary/goal of the currently highlighted session as you navigate with `↑`/`↓` (`k`/`j`), with a windowed list view preventing viewport overflow.

---

## 8. Host Servers & Plugins

MCP servers and enabled plugins in your normal agent config (the host) are available in
every aim session. aim merges them into the profile when a session starts and takes them
out again when the last session of that profile exits, so at rest a profile holds only
its own. What is merged:

| Agent | MCP servers | Plugins |
|---|---|---|
| Claude | `mcpServers` in `~/.claude.json` | `enabledPlugins` and `extraKnownMarketplaces` in `~/.claude/settings.json` |
| Codex | `[mcp_servers.*]` in `~/.codex/config.toml` | `[plugins."<id>"]` in the same file, with the plugin's own sub-tables |
| agy | `mcpServers` in the shared `mcp_config.json` | — (agy's plugins are fully shared already) |

Only which plugins are on is scoped per profile; the installed plugin files and
marketplaces stay shared with the host.

When a session adds, edits or removes a server or plugin, aim asks at exit whether to
promote the change to the host or keep it in that profile. When stdin is piped
(`aim run claude work < file`) the prompt uses the controlling terminal (`/dev/tty`);
without a terminal at all, or on Ctrl+C or a closed terminal at the prompt, the change is kept. A backgrounded run (`aim run … &`) keeps every change without prompting. Each change is listed with its collection, and a plugin switched on or off shows
its new value (`~ enabledPlugins/x@m   edited (host item) → false`). A removed host item
comes back next session. Background launches (`claude --bg`, `codex app-server`,
`agy remote-control`) get the host items too; the next launch of that profile cleans up.
Version and help invocations skip the merge entirely, as each CLI spells them: for
`claude`, `-v`, `-V`, `--version`, `-h` and `--help`; for `codex`, `-V`, `--version`, `-h`,
`--help` and `codex help`; for `agy`, `--version`, `-version`, `-h`, `--help`, `-help` and
`agy help`. Anything else, such as `claude version` (a prompt), starts a session. Skipping
the merge also leaves cleanup of a crashed or background session to the next real session.

To switch a host plugin off in one profile only, disable it inside a session and keep the
change: `aim run claude work plugin disable x@m` (for Codex, set `enabled = false` in the
profile's `config.toml`). The profile's `false` is its own from then on and wins over the
host's `true` in every later session. An item the profile defines always wins over the
host's.

To add a server or plugin to one profile only, run the agent's own command through aim and
keep the change at exit: `aim run claude work mcp add -s user foo -- npx foo`.

The first launch of an existing profile removes servers and plugins that are identical
copies of the host's. Before it does, aim writes a backup next to the profile file, named
`<file>.aim-backup-<UTC timestamp>`; delete those backups once the profile looks right. A
profile entry that differs from the host's stays the profile's own. For example, a Codex
profile holding 17 plugin tables identical to the host's and one the host no longer has
loses the 17 (backed up), keeps the other as its own, and sees the host's plugins in its
sessions. A new profile starts with none of the host's servers or plugins in its config.

While a foreground `aim run` session of a profile is live, `aim remove`, `aim mv` and the TUI's
rename, move and delete refuse with "profile <p> has a running <agent> session; exit it
first". Background launches (`claude --bg`, `codex app-server`,
`agy remote-control`) are not detected.

Agents started by hand in a profile, outside `aim run` (an IDE pointed at the profile's
`CLAUDE_CONFIG_DIR` or `CODEX_HOME`), see only the profile's own servers and plugins.

Turn the merge off per profile in `~/.aim/config.json`: `"mcp_global": false` for servers,
`"plugins_global": false` for plugins.

---

## 9. Troubleshooting & Diagnostics

For common operational issues, edge cases, and step-by-step remedies:
- OAuth refresh token revocation (`Your access token could not be refreshed...`)
- Cross-profile session resumption and 0-turn history recovery
- Codex SQLite migration collisions (`table threads already exists`)
- TOML syntax & duplicate key parse errors
- Caveman compression proxy sidecar connectivity
- Skill descriptions context shortening notices
- macOS Keychain isolation & developer tool authentication (`git`, `gh`)
- Gatekeeper quarantine clearance

See the complete **[Troubleshooting Guide](TROUBLESHOOTING.md)**.

---

## 10. In-Tool Feedback & Crash Reporting

AIM includes direct feedback channels to report issues, suggest features, or report bugs without context switching:

### 10.1 CLI Feedback (`aim feedback`)
Submit feedback from the terminal:
```bash
aim feedback "Love the Codex session resumption support!"
aim feedback --category bug "Quota reset timer displays negative offset on macOS"
aim feedback --include-doctor   # attaches an anonymized doctor diagnostic report
```
If piped:
```bash
echo "Feature idea: add Claude Code tool call analytics" | aim feedback
```
Standard input is bounded to 32KB to protect system resources.

### 10.2 TUI Feedback Modal (`F`)
Pressing `F` in the interactive dashboard opens the feedback modal:
- Cycle category with `Tab` (General Feedback 💬, Bug Report 🐛, Feature Request 💡).
- Toggle inclusion of the anonymized diagnostic report with `Ctrl+D`.
- Submit with `Enter`.

### 10.3 Panic Recovery & Sanitized Crash Bundles
AIM registers a top-level crash recovery boundary across all command invocations. If an unexpected panic occurs:
- The panic message and stack trace are scrubbed of personal home paths (`~`) and credentials (OAuth tokens, API keys, Bearer headers).
- A structured crash report is written to disk at `~/.aim/reports/`.
- A pre-filled GitHub issue URL (capped at 4,000 characters to prevent HTTP 414 errors) is printed to the terminal for one-click issue creation.

---

## 11. Privacy & Anonymous Telemetry

AIM implements privacy-by-design anonymous telemetry to understand command performance, failure rates, and agent adoption.

### 11.1 Anonymity Guarantees
- **No Personal Identifiers**: Machine IDs are random SHA-256 hashes generated locally with `0600` permissions (`~/.aim/telemetry_id`). They contain no usernames, hostnames, MAC addresses, or IP addresses.
- **Strict Sanitization**: Command names are matched against a known whitelist (`run`, `list`, `doctor`, `sessions`, `feedback`, `whoami`). User-supplied arguments, prompts, profile names, repository names, and flags are stripped before emission.
- **Credential Redaction**: Stack traces, error logs, and reports pass through regex scrubbers redacting Google (`ya29.`), GitHub (`ghp_`), OpenAI (`sk-`), PostHog (`phc_`/`phx_`), and Bearer authorization tokens.
- **Zero Ingestion Leakage**: PostHog ingestion rules and client-side event whitelisting drop any unauthorized event names.

### 11.2 Opt-Out Precedence
Telemetry collection can be fully disabled at any time using standard conventions:
1. `export DO_NOT_TRACK=1` (universal web/tool standard)
2. `export AIM_TELEMETRY_DISABLED=1`
3. `"telemetry": false` in `~/.aim/config.json`

When any opt-out tier is active, telemetry spooling is completely bypassed and zero bytes are written to disk or sent over the network. Test suites run with strict test isolation to prevent telemetry emission in development.



