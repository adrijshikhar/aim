# AIM (AI Multiplexer) — Troubleshooting Guide

This guide provides actionable diagnosis and resolution steps for real-world issues encountered when multiplexing AI coding assistants (Antigravity CLI, OpenAI Codex CLI, Gemini CLI) across isolated profiles with `aim`.

---

## Table of Contents

- [Quick Diagnostics Toolkit](#quick-diagnostics-toolkit)
- [1. Authentication & OAuth Issues](#1-authentication--oauth-issues)
  - [Revoked or Expired OAuth Refresh Token](#revoked-or-expired-oauth-refresh-token)
  - [Accidental Profile Creation via Typo Prompting Re-Login (`office--yolo`)](#accidental-profile-creation-via-typo-prompting-re-login-office--yolo)
  - [OAuth Browser Does Not Open Automatically](#oauth-browser-does-not-open-automatically)
  - [Antigravity Access Token Expired](#antigravity-access-token-expired)
- [2. Session Resumption & Cross-Profile History](#2-session-resumption--cross-profile-history)
  - [Session Not Listed in Codex Native `/resume` Picker](#session-not-listed-in-codex-native-resume-picker)
  - [Passing Custom Agent Flags on Resume (`--model`, `--sandbox`, etc.)](#passing-custom-agent-flags-on-resume---model---sandbox-etc)
  - [Resumed Session Has 0 Turns or Missing Conversation History](#resumed-session-has-0-turns-or-missing-conversation-history)
  - [Compacted / Continuation Session Loses Context](#compacted--continuation-session-loses-context)
  - [Ambiguous Session ID Prefix Error](#ambiguous-session-id-prefix-error)
- [3. Database & SQLite Migration Collisions](#3-database--sqlite-migration-collisions)
  - [Codex `sqlx` Migration 1 Conflict (`table threads already exists`)](#codex-sqlx-migration-1-conflict-table-threads-already-exists)
  - [Malformed SQLite Database or Locked Database (`busy / locked`)](#malformed-sqlite-database-or-locked-database-busy--locked)
- [4. Configuration & TOML Parse Errors](#4-configuration--toml-parse-errors)
  - [Duplicate Table Key in `config.toml`](#duplicate-table-key-in-configtoml)
  - [MCP Server Collisions Across Profiles](#mcp-server-collisions-across-profiles)
- [5. Sidecar & Proxy Connectivity (Caveman Mode)](#5-sidecar--proxy-connectivity-caveman-mode)
  - [Connection Refused on `127.0.0.1:8787`](#connection-refused-on-1270018787)
  - [Caveman Proxy Request Timeout](#caveman-proxy-request-timeout)
- [6. Skill Context Budget Warnings](#6-skill-context-budget-warnings)
  - [Yellow Notice: `Skill descriptions were shortened to fit the skills context budget`](#yellow-notice-skill-descriptions-were-shortened-to-fit-the-skills-context-budget)
- [7. macOS Keychain Isolation & Developer Credentials](#7-macos-keychain-isolation--developer-credentials)
  - [Git or GitHub CLI (`gh`) Prompts for Login Inside Subshell](#git-or-github-cli-gh-prompts-for-login-inside-subshell)
  - [Agent Credentials Leaking into macOS Keychain](#agent-credentials-leaking-into-macos-keychain)
- [8. Terminal Ghost Sessions & Zombie PIDs](#8-terminal-ghost-sessions--zombie-pids)
  - [Session Remains Marked `[ACTIVE]` After Window Closed](#session-remains-marked-active-after-window-closed)
- [9. macOS Gatekeeper Quarantine](#9-macos-gatekeeper-quarantine)
  - [`"aim" cannot be opened because the developer cannot be verified`](#aim-cannot-be-opened-because-the-developer-cannot-be-verified)
- [Summary Diagnostics Matrix](#summary-diagnostics-matrix)

---

## Quick Diagnostics Toolkit

Whenever you encounter unexpected behavior, run these first:

```bash
# 1. Inspect environment, tokens, binaries, and dotfile links
aim doctor

# 2. Check targeted agent diagnostics (e.g. codex, agy, gemini)
aim doctor codex

# 3. View your active profile, agent, session, and quota telemetry
aim whoami

# 4. Run any command with verbose debug logging to stderr and ~/.aim/aim-debug.log
aim --debug run codex work
aim --debug resume codex work <session-id>
```

Debug logs are permanently recorded at:
```text
~/.aim/aim-debug.log
```

---

## 1. Authentication & OAuth Issues

### Revoked or Expired OAuth Refresh Token

#### Symptom
When starting Codex, the session terminates immediately or prints:
```text
■ Your access token could not be refreshed because your refresh token was revoked. Please log out and sign in again.
```

#### Root Cause
OpenAI OAuth refresh tokens expire or are revoked when:
- Password, security keys, or session tokens are changed on ChatGPT/OpenAI.
- A concurrent login from another client invalidated the refresh token.
- The refresh token timed out after prolonged inactivity.

Because AIM isolates credentials inside `~/.aim/profiles/<profile>/.codex/auth.json`, logging in on the host or in another profile will **not** update this profile's token.

#### Resolution
Re-authenticate the specific profile directly using `aim login`:
```bash
aim login codex <profile>
# Example:
aim login codex work
```
This triggers the browser-based OAuth PKCE flow and writes new tokens directly into the target profile's isolated `auth.json`. Verify the fix with:
```bash
aim doctor codex
```

---

### Accidental Profile Creation via Typo Prompting Re-Login (`office--yolo`)

#### Symptom
You run a command like `aim run codex office--yolo`, and unexpectedly Codex prompts you to sign in with ChatGPT again. After signing in, there are zero past sessions to resume from, making it appear as if your original profile credentials and conversation history were wiped.

#### Root Cause
Missing a space before a flag (e.g. typing `office--yolo` instead of `office --yolo` or `office -- --yolo`) causes the shell and CLI parser to treat the whole string as a brand-new profile name (`"office--yolo"`).
Because this profile never existed before:
1. It has no `auth.json`, triggering a fresh login prompt.
2. It has an empty SQLite database with 0 past sessions.

Your real profile (e.g. `office` or `work`) is completely safe and untouched.

#### AIM Safeguard
AIM now validates profile existence before launching:
- **Interactive Prompt**: If the profile does not exist, AIM stops and asks:
  ```text
  Warning: profile name "office--yolo" contains "--". Did you mean "office" with flag "--yolo"?
  Profile "office--yolo" does not exist. Do you want to create it and start codex? [y/N]:
  ```
  Pressing Enter or `N` aborts cleanly without creating any phantom directory.
- **Non-Interactive Guard**: In scripts or non-TTY environments, AIM immediately exits with an error rather than creating an unauthenticated sandbox.

#### Recovery & Cleanup
1. Remove the accidental profile:
   ```bash
   aim remove codex office--yolo
   # or
   aim remove office--yolo
   ```
2. Launch your real profile:
   ```bash
   aim run codex office
   # or resume your previous session:
   aim resume codex office <session-id>
   ```
3. To pass flags to Codex, always include a space and the `--` separator:
   ```bash
   aim run codex office -- --dangerously-bypass-approvals-and-sandbox
   ```

---

### OAuth Browser Does Not Open Automatically

#### Symptom
Running `aim login agy <profile>` or `aim login codex <profile>` hangs at `Waiting for authentication...` but no browser tab opens.

#### Root Cause
On headless machines, remote SSH sessions, or Linux environments lacking `xdg-open` / macOS `open`, the automatic browser invocation cannot launch GUI applications.

#### Resolution
AIM outputs the full authorization URL directly to the terminal:
1. Copy the displayed URL from your terminal.
2. Open it in any browser where you are logged into your account.
3. Complete the authorization; the local loopback listener (e.g., `http://localhost:14555/callback`) receives the token automatically.

---

### Antigravity Access Token Expired

#### Symptom
Antigravity CLI commands fail with `HTTP 401 Unauthorized` or `Invalid OAuth token`.

#### Root Cause
Antigravity CLI short-lived access tokens expire every 60 minutes. Usually, the token refresher automatically renews it; if offline or if the refresh token expired, renewal fails.

#### Resolution
Run `aim doctor agy` to check the remaining token lifetime:
```bash
aim doctor agy
```
If expired or missing, re-authenticate:
```bash
aim login agy <profile>
```

---

## 2. Session Resumption & Cross-Profile History

### Session Not Listed in Codex Native `/resume` Picker

#### Symptom
You worked in the `office` profile, switched to the `work` profile, and typed `/resume` inside Codex, but the previous session does not appear.

#### Root Cause
Agent CLIs are fully sandboxed. Codex's internal `/resume` command only inspects its local SQLite database (`$CODEX_HOME/state_5.sqlite`) and its local rollout directory (`$CODEX_HOME/sessions/`). It has no awareness of other AIM profiles.

#### Resolution
Use AIM's built-in session resumption, which bridges sessions across profiles:
```bash
# 1. Find the session ID across all profiles:
aim sessions codex

# 2. Resume the session inside your target profile:
aim resume codex work <session-id>
```
Alternatively, in the TUI (`aim`):
1. Press `s` to open the **Sessions Explorer**.
2. Select the session you want.
3. Press `Enter` to resume verbatim, or `c` for a Catalyst handoff.

AIM's hydrator automatically clones the thread metadata, ancestor rollouts, and turn history into the target profile before launching.

---

### Invalid Paginated History Lineage: Cutoff Byte Offset Past Source Rollout

#### Symptom
When resuming a session that was created as a continuation/compacted thread from an earlier session, Codex fails during bootstrap with:
```text
Error: Failed to resume session from .../rollout-...jsonl: thread/resume failed during TUI bootstrap: thread/resume failed: invalid paginated history lineage for <session-id>: cutoff byte offset is past the source rollout (code -32600)
```

#### Root Cause
Codex uses `paginated` history mode for compacted sessions, where a child rollout contains a `history_base` reference pointing to an ancestor thread ID and an `end_byte_offset`. If the ancestor session was previously copied into the target profile at an earlier point when it had fewer turns, the target profile's copy of the ancestor rollout file is smaller than `end_byte_offset`. When Codex attempts to read the ancestor rollout up to that byte offset, it fails with code `-32600`.

#### Resolution
AIM automatically detects when a destination ancestor rollout is smaller or older than the source ancestor rollout and refreshes it with the complete file during `aim resume` and `aim sessions import`.

If you encounter this manually, re-import the session to refresh all ancestor trees:
```bash
aim sessions import codex <target-profile> <session-id>
```

---


### Passing Custom Agent Flags on Resume (`--model`, `--sandbox`, etc.)

#### Symptom
When attempting to pass agent flags directly to `aim resume`, you receive an unknown flag error:
```bash
aim resume codex work 01a0a581 -m gpt-5.6-terra
# Error: unknown shorthand flag: 'm' in -m

aim resume codex work 01a0a581 --search
# Error: unknown flag: --search
```

#### Root Cause
`aim resume` defines its own CLI flags (`--exact`, `-c`/`--catalyst`, `-b`/`--fork`, `-f`/`--force`). By default, the CLI argument parser interprets all preceding flags as AIM options. Agent-specific flags (like Codex's `-m`, `-s`, `--search`, or Antigravity's `--mode`) are not recognized by AIM's top-level parser.

#### Resolution: The Double-Dash (`--`) Delimiter
To pass flags and arguments directly to the underlying agent binary (`codex`, `agy`), use the POSIX standard **`--`** separator. Everything after `--` is forwarded verbatim to the agent CLI:

```bash
aim resume <agent> <profile> [session-id] [aim-flags] -- [agent-flags...]
```

#### Common Examples

##### 1. Override the LLM Model on Resume
```bash
# Resume with a specific model (e.g. gpt-5.6-terra, gpt-6-astra, o3):
aim resume codex work 01a0a581 -- -m gpt-5.6-terra
```

##### 2. Change the Execution Sandbox Policy
```bash
# Allow workspace file writes:
aim resume codex work 01a0a581 -- -s workspace-write

# Unrestricted execution mode (danger-full-access):
aim resume codex work 01a0a581 -- -s danger-full-access
```

##### 3. Enable Live Web Search
```bash
aim resume codex work 01a0a581 -- --search
```

##### 4. Skip or Customize Human Approvals
```bash
# Bypass all approval prompts:
aim resume codex work 01a0a581 -- --dangerously-bypass-approvals-and-sandbox

# Never prompt for confirmation:
aim resume codex work 01a0a581 -- -a never
```

##### 5. Provide an Initial Continuation Prompt Directly
```bash
aim resume codex work 01a0a581 -- "Continue refactoring and run make test"
```

##### 6. Combine AIM Flags with Agent Flags
You can combine AIM flags (`--fork`, `--force`, `--catalyst`) with agent flags separated by `--`:
```bash
# Fork conversation into a new ID AND switch models:
aim resume codex work 01a0a581 --fork -- -m gpt-5.6-terra

# Force resume an active session with full access:
aim resume codex work 01a0a581 --force -- -s danger-full-access

# Resume via Catalyst handoff with custom model:
aim resume codex work 01a0a581 --catalyst -- -m gpt-6-astra
```

##### 7. Antigravity CLI Custom Flags
```bash
# Resume an Antigravity conversation in plan mode:
aim resume agy work 775e6ada -- --mode=plan
```

---

### Resumed Session Has 0 Turns or Missing Conversation History

#### Symptom
You resumed a Codex session across profiles, but Codex shows `0 turns` or starts with an empty prompt history.

#### Root Cause
Codex CLI (v0.154+) splits session data between two databases:
1. `state_5.sqlite` (thread metadata, `threads` table).
2. `thread_history_1.sqlite` (granular conversation turns: `thread_turns`, `thread_items`, and projection states).

If only `state_5.sqlite` was migrated, the thread exists but has zero turns.

#### Resolution
Ensure you are running AIM v0.4.0 or newer, which includes full-fidelity multi-database hydration. Re-import or resume the session:
```bash
# Hydrate and launch with complete turn fidelity:
aim resume codex work <session-id>

# Or manually hydrate from host/another profile:
aim sessions import codex work <session-id>
```

---

### Compacted / Continuation Session Loses Context

#### Symptom
Resuming a compacted Codex session fails with `parent thread not found` or misses earlier context.

#### Root Cause
When Codex compacts context or continues a session, it creates a new rollout file whose header references the parent session:
```json
{"history_base":{"thread_id":"0198a...","current_turn_index":14}}
```
Resuming the child session without also copying the ancestor rollout and parent database rows breaks the lineage chain.

#### Resolution
AIM's hydrator automatically traverses the entire ancestor tree recursively:
```bash
aim resume codex work <child-session-id>
```
AIM detects `history_base.thread_id`, resolves all parent rollout files and ancestor rows in `state_5.sqlite` and `thread_history_1.sqlite`, and copies them into the target profile in proper topological order.

---

### Ambiguous Session ID Prefix Error

#### Symptom
Running `aim resume codex work 01a0` fails with:
```text
Error: ambiguous session ID prefix "01a0" matches multiple sessions (01a0a581, 01a0b92c). Provide more characters.
```

#### Root Cause
Session ID short prefixes must resolve to exactly one unique conversation thread.

#### Resolution
Provide at least 6–8 characters of the session ID:
```bash
aim resume codex work 01a0a581
```

---

## 3. Database & SQLite Migration Collisions

### Codex `sqlx` Migration 1 Conflict (`table threads already exists`)

#### Symptom
Codex exits immediately on startup with:
```text
Error: migration 1 failed: table threads already exists
```

#### Root Cause
Codex uses Rust's `sqlx` migration engine. Migration state is tracked in `_sqlx_migrations`. If a third-party tool or manual script created `CREATE TABLE threads` without recording the migration entries, `sqlx` panics upon startup.

#### Resolution
AIM automatically prevents this by initializing fresh databases using SQLite's native `VACUUM INTO` from a known valid schema and truncating table rows cleanly.

To manually recover a corrupted profile database:
```bash
# 1. Open an isolated subshell for the profile:
aim shell codex <profile>

# 2. Back up and remove the conflicting SQLite files:
mv ~/.codex/state_5.sqlite ~/.codex/state_5.sqlite.bak
mv ~/.codex/thread_history_1.sqlite ~/.codex/thread_history_1.sqlite.bak

# 3. Launch codex once to allow sqlx to run clean migrations:
codex --help
exit
```
Your authentication in `auth.json` remains untouched.

---

### Malformed SQLite Database or Locked Database (`busy / locked`)

#### Symptom
```text
Error: database disk image is malformed
# or
Error: database is locked (busy)
```

#### Root Cause
Abrupt machine power-off, SIGKILL, or multiple concurrent processes accessing the same un-isolated profile can corrupt SQLite WAL journals or hold locks.

#### Resolution
1. Verify no zombie processes are holding locks:
   ```bash
   aim sessions codex --active
   ```
2. Remove stale WAL and shared memory lock files:
   ```bash
   rm -f ~/.aim/profiles/<profile>/.codex/*.sqlite-wal
   rm -f ~/.aim/profiles/<profile>/.codex/*.sqlite-shm
   ```
3. Run `aim doctor codex` to verify database health.

---

## 4. Configuration & TOML Parse Errors

### Duplicate Table Key in `config.toml`

#### Symptom
Codex fails to start with:
```text
Error: toml: duplicate table key: mcp_servers.<name>
# or
Error: duplicate key in codex config
```

#### Root Cause
Automated installer scripts or manual edits may append duplicate `[mcp_servers.<name>]` headers to `~/.aim/profiles/<profile>/.codex/config.toml`. The TOML specification strictly forbids duplicate tables.

#### Resolution
1. Inspect the profile's configuration:
   ```bash
   cat ~/.aim/profiles/<profile>/.codex/config.toml
   ```
2. Search for repeated section headers:
   ```bash
   grep '^\[' ~/.aim/profiles/<profile>/.codex/config.toml | sort | uniq -d
   ```
3. Edit the file to merge duplicate sections into a single definition.

---

### MCP Server Collisions Across Profiles

#### Symptom
An MCP server works in one profile but fails with `connection refused` or `command not found` in another.

#### Root Cause
Profiles virtualize `$HOME`. If an MCP server in `config.toml` uses a relative path (e.g. `node ./server.js` or `~/bin/mcp-server`), the relative path resolves relative to the *profile's* virtual home directory, where the binary may not exist.

#### Resolution
Always specify **absolute paths** or global system paths for MCP executable commands in `config.toml`:
```toml
[mcp_servers.playwright]
command = "/usr/local/bin/npx"
args = ["-y", "@playwright/mcp-server"]
```

---

## 5. Sidecar & Proxy Connectivity (Caveman Mode)

### Connection Refused on `127.0.0.1:8787`

#### Symptom
When running Codex with Caveman compression enabled (`model_provider = "caveman"`), queries fail with:
```text
Error: connection refused: 127.0.0.1:8787
```

#### Root Cause
The Caveman proxy sidecar daemon is either not running or failed to bind to port 8787.

#### Resolution
1. Check if the Caveman daemon process is alive:
   ```bash
   ps aux | grep caveman
   curl -s http://127.0.0.1:8787/health
   ```
2. If using AIM's automated runner, AIM detects `caveman` in `config.toml` and launches the sidecar automatically. Run with `--debug` to inspect the sidecar lifecycle:
   ```bash
   aim --debug run codex <profile>
   ```
3. If you do not intend to use the Caveman proxy, switch `model_provider` back to `openai` in `~/.aim/profiles/<profile>/.codex/config.toml`.

---

### Caveman Proxy Request Timeout

#### Symptom
Codex hangs during generation when routed through the Caveman proxy.

#### Root Cause
Upstream LLM rate limiting or streaming buffer stall between the sidecar and the model provider.

#### Resolution
Check the sidecar logs or test upstream connectivity directly. You can bypass the sidecar on a single run by passing the direct model flag:
```bash
aim run codex <profile> -- -m gpt-5.6-terra
```

---

## 6. Skill Context Budget Warnings

### Yellow Notice: `Skill descriptions were shortened to fit the skills context budget`

#### Symptom
When starting Codex, a yellow warning banner appears:
```text
Skill descriptions were shortened to fit the skills context budget.
```

#### Root Cause
Codex reserves a fixed token window (typically ~1–2% of the initial context window) for indexing available skill descriptions. When you have dozens of skills installed in `.agents/skills/` or `~/.codex/skills/`, Codex automatically trims description strings to stay within budget.

#### Impact & Resolution
> [!NOTE]
> **This warning is harmless and non-breaking.**
> All installed skills remain fully functional. When a skill is invoked explicitly (e.g. via `$skill-name`) or triggered by relevant keywords, Codex loads the full `SKILL.md` dynamically into context.

To silence the warning:
1. Audit unused skills in `.agents/skills/` and remove those you don't use.
2. Shorten the `description` field in the frontmatter of your custom `SKILL.md` files.

---

## 7. macOS Keychain Isolation & Developer Credentials

### Git or GitHub CLI (`gh`) Prompts for Login Inside Subshell

#### Symptom
Running `git push`, `git fetch`, or `gh pr create` inside `aim shell` or an agent session prompts for credentials or fails with:
```text
fatal: could not read Username for 'https://github.com': terminal prompts disabled
```

#### Root Cause
AIM redirects `$HOME` to `~/.aim/profiles/<profile>`. By default, AIM bridges `~/.gitconfig`, `~/.ssh/`, and the macOS `Library/Keychains` so developer credentials pass through seamlessly.
However, if Git is configured to store credentials in a file inside `$HOME` (e.g. `~/.git-credentials`) rather than the macOS keychain or SSH agent, the profile sandbox will not see it.

#### Resolution
1. Verify `git` is configured to use the macOS keychain:
   ```bash
   git config --global credential.helper osxkeychain
   ```
2. Verify SSH keys are loaded into your SSH agent:
   ```bash
   ssh-add -l
   # If empty, add your key:
   ssh-add ~/.ssh/id_ed25519
   ```
3. If using `gh`, run `gh auth status` on the host, or run `gh auth login` once inside `aim shell <agent> <profile>`.

---

### Agent Credentials Leaking into macOS Keychain

#### Symptom
Switching between profiles unexpectedly uses the other profile's API key or account.

#### Root Cause
Certain agent binaries (such as older Antigravity or Gemini CLI builds) attempt to write authentication tokens to the default macOS login keychain (`login.keychain-db`).

#### Resolution
AIM enforces an **Explicit Agent Ignore List**:
1. Run `aim doctor`:
   ```bash
   aim doctor
   ```
   AIM scans the keychain, detects any agent tokens, and displays:
   ```text
   [WARN] Keychain: Agent token(s) detected in macOS Keychain (potential profile isolation risk)
          Auto-purging ignored agent keychains to enforce profile isolation...
   [OK]   Keychain: Agent credentials successfully purged from macOS Keychain.
   ```
2. This forces the agent to read and write tokens exclusively within its sandboxed virtual home directory.

---

## 8. Terminal Ghost Sessions & Zombie PIDs

### Session Remains Marked `[ACTIVE]` After Window Closed

#### Symptom
`aim sessions` reports a session as `[ACTIVE]` with a green indicator even though the terminal window was closed.

#### Root Cause
If the terminal window was force-quit (SIGKILL, macOS window crash, or machine reboot), the agent process may not have executed its exit handlers.

#### Resolution
AIM's process scanner automatically checks the OS process table for live PIDs:
1. Run `aim sessions`:
   ```bash
   aim sessions
   ```
   AIM inspects the PID list, detects that the process is no longer alive, and updates the session index to mark it as exited.
2. If a background process is truly still running and hung:
   ```bash
   # Find and terminate the hung process:
   ps aux | grep codex
   kill -9 <PID>
   ```

---

## 9. macOS Gatekeeper Quarantine

### `"aim" cannot be opened because the developer cannot be verified`

#### Symptom
When executing `aim` after downloading the binary directly from GitHub Releases in a browser, macOS displays a security popup blocking execution.

#### Root Cause
macOS applies the `com.apple.quarantine` extended attribute to all binaries downloaded via web browsers.

#### Resolution
Clear the quarantine attribute via terminal:
```bash
xattr -d com.apple.quarantine $(which aim)
```

> [!TIP]
> **Preferred Installation Method**: Install via Homebrew to prevent Gatekeeper warnings entirely:
> ```bash
> brew tap adrijshikhar/homebrew-tap
> brew install --cask aim
> ```

---

## Summary Diagnostics Matrix

| Symptom / Error | Primary Cause | Immediate Fix |
|---|---|---|
| `refresh token was revoked` | OAuth refresh token expired/revoked | `aim login codex <profile>` |
| Asked to sign in & 0 sessions on resume | Accidental profile created via typo (e.g. `office--yolo`) | `aim remove <typo-profile>` & run real profile |
| Session missing in `/resume` | Cross-profile sandbox isolation | `aim resume <agent> <profile> <id>` |
| `unknown shorthand flag` on resume | Missing `--` delimiter | `aim resume <agent> <profile> <id> -- [flags...]` |
| Resumed session has 0 turns | Missing `thread_history_1.sqlite` rows | Upgrade to AIM ≥ 0.4.0 & re-resume |
| `migration 1 failed: table threads already exists` | `sqlx` migration version collision | `aim doctor codex` / clean DB init |
| `duplicate table key: mcp_servers...` | Duplicate TOML headers in `config.toml` | Deduplicate headers in `config.toml` |
| `connection refused 127.0.0.1:8787` | Caveman compression sidecar offline | Check `ps aux \| grep caveman` or reset `model_provider` |
| `Skill descriptions were shortened...` | Codex skills token budget reached | Harmless warning; prune unused skills if desired |
| `fatal: could not read Username` in git | Git credential helper not set to osxkeychain | `git config --global credential.helper osxkeychain` |
| macOS Gatekeeper popup | `com.apple.quarantine` flag | `xattr -d com.apple.quarantine $(which aim)` |

---

## Related Documentation

- 📖 [Wiki & Architecture Guide](file:///Users/nemesis/Projects/my-projects/aim/wiki.md) — Virtual home architecture, isolation boundaries, and configuration schema.
- 🎨 [Design System](file:///Users/nemesis/Projects/my-projects/aim/design.md) — Atom One Dark color palette, UI tokens, and Lipgloss standards.
- 🛠️ [Contributing Guide](file:///Users/nemesis/Projects/my-projects/aim/CONTRIBUTING.md) — Local development, test suite execution, and agent adapters.
- 🚀 [README](file:///Users/nemesis/Projects/my-projects/aim/README.md) — Installation, quickstart, and CLI cheat sheet.
