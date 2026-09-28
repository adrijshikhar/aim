#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# Parameterized E2E Integration Test Suite for AIM
# Tests all registered agent adapters (agy, codex, claude, gemini, and future additions)
#
# Coverage:
# 1. Environment Isolation (HOME, AIM_AGENT, AIM_PROFILE, storage dirs)
# 2. Arbitrary Flag Forwarding (no Cobra swallowing of --dangerously-skip-permissions, etc.)
# 3. Typo Safety Guards (refusal to silently create profiles containing '--', interactive suggestions)
# 4. Profile Dotfiles Bridging (.gitconfig, .ssh)
# 5. Diagnostic Reporting (aim doctor)
# 6. Session Listing across profiles (aim sessions)
# 7. Cross-Profile Session Hydration & Flag Forwarding (aim resume)
# 8. Size-Aware Deduplication & Sync (larger source rollout overwrites stale/truncated touch)
# 9. Session Forking (--fork with UUID regeneration and ref replacement)
# 10. Session-Scoped MCP Server Merge (host servers merged per session, keep / promote,
#     mcp_global:false, background launch (CLI or profile args) + recovery; adapters
#     without MCP untouched)
# ==============================================================================

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AIM_BIN="$REPO_ROOT/aim"
echo "==> Building fresh AIM binary from $REPO_ROOT..."
(cd "$REPO_ROOT" && go build -o "$AIM_BIN" ./cmd/aim)

TEST_DIR=$(mktemp -d)
MOCK_BIN=$(mktemp -d)
trap 'rm -rf "$TEST_DIR" "$MOCK_BIN"' EXIT

export AIM_HOME="$TEST_DIR/aim_home"
export AIM_REAL_HOME="$TEST_DIR/real_home"
export PATH="$MOCK_BIN:$PATH"

mkdir -p "$AIM_REAL_HOME/.ssh"
touch "$AIM_REAL_HOME/.gitconfig"
mkdir -p "$AIM_REAL_HOME/.gemini/antigravity-cli"

# Mock macOS security binary to isolate Keychain during tests
cat << 'MOCK' > "$MOCK_BIN/security"
#!/bin/sh
exit 0
MOCK
chmod +x "$MOCK_BIN/security"

# ==============================================================================
# Adapter Definition Matrix
# To onboard a new adapter (e.g. 'opencode', 'copilot'), simply append to ALL_ADAPTERS
# and define its mock binary and session hooks below.
# ==============================================================================
ALL_ADAPTERS=("agy" "codex" "claude" "gemini")

# Mock Agent Binaries
for AGENT in "${ALL_ADAPTERS[@]}"; do
  cat << MOCK > "$MOCK_BIN/$AGENT"
#!/bin/sh
if [ "\$1" = "--version" ]; then
  echo "$AGENT-cli 1.0.0-test"
  exit 0
fi
if [ "\$1" = "auth" ] && [ "\$2" = "login" ]; then
  echo "mock $AGENT login successful"
  exit 0
fi
echo "AGENT_INVOKED: $AGENT" >> "$TEST_DIR/${AGENT}_invoked.log"
echo "ARGS: \$*" >> "$TEST_DIR/${AGENT}_invoked.log"
echo "HOME: \$HOME" >> "$TEST_DIR/${AGENT}_invoked.log"
echo "AIM_AGENT: \${AIM_AGENT:-}" >> "$TEST_DIR/${AGENT}_invoked.log"
echo "AIM_PROFILE: \${AIM_PROFILE:-}" >> "$TEST_DIR/${AGENT}_invoked.log"
# E2E_HOOK simulates work the agent does mid-session (e.g. a native `mcp add`)
if [ -n "\${E2E_HOOK:-}" ]; then sh "\$E2E_HOOK"; fi
MOCK
  chmod +x "$MOCK_BIN/$AGENT"
done

# Helper: Check if adapter supports session management
adapter_has_sessions() {
  case "$1" in
    agy|codex|claude) return 0 ;;
    *) return 1 ;;
  esac
}

# Helper: Check if adapter supports session forking
adapter_supports_fork() {
  case "$1" in
    codex|claude) return 0 ;;
    *) return 1 ;;
  esac
}

# Helper: Setup multi-turn test session in source profile
setup_test_session() {
  local agent="$1"
  local profile="$2"
  local session_id="$3"
  local prof_dir="$AIM_HOME/profiles/$profile"

  case "$agent" in
    codex)
      local sess_dir="$prof_dir/.codex/sessions/2026/09/22"
      local shell_dir="$prof_dir/.codex/shell_snapshots"
      mkdir -p "$sess_dir" "$shell_dir"

      # Write rich 100-turn parent rollout
      python3 - "$sess_dir/rollout-$session_id.jsonl" << 'PYEOF'
import sys
with open(sys.argv[1], "w") as f:
    for i in range(100):
        f.write(f'{{"event":"turn","turn_index":{i},"content":"turn content {i} with extensive tooling work"}}\n')
PYEOF
      # Child subagents
      echo '{"event":"subagent_luna_init"}' > "$sess_dir/rollout-01a0ce1e-af6b-76d1-a39c-84394870f6f3.jsonl"
      echo '{"event":"subagent_astra_init"}' > "$sess_dir/rollout-01a0ce22-d3e0-79d0-b8ae-cce0dd14fa20.jsonl"
      echo '{"shell_snapshot":"active_term"}' > "$shell_dir/snapshot-1.json"

      # SQLite database with threads and spawn edges
      local db="$prof_dir/.codex/state_5.sqlite"
      sqlite3 "$db" <<SQL
CREATE TABLE IF NOT EXISTS _sqlx_migrations (version BIGINT PRIMARY KEY, description TEXT NOT NULL, installed_on TIMESTAMP, success BOOLEAN, checksum BLOB, execution_time BIGINT);
INSERT OR IGNORE INTO _sqlx_migrations VALUES (1, 'init', CURRENT_TIMESTAMP, 1, X'00', 1);
CREATE TABLE IF NOT EXISTS threads (
	id TEXT PRIMARY KEY,
	rollout_path TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	source TEXT NOT NULL,
	model_provider TEXT NOT NULL,
	cwd TEXT NOT NULL,
	title TEXT NOT NULL,
	sandbox_policy TEXT NOT NULL,
	approval_mode TEXT NOT NULL
);
INSERT OR REPLACE INTO threads VALUES
('$session_id', '$sess_dir/rollout-$session_id.jsonl', 1727100000, 1727100000, 'cli', 'openai', '/workspace', 'codex-e2e-task', 'danger', 'manual'),
('01a0ce1e-af6b-76d1-a39c-84394870f6f3', '$sess_dir/rollout-01a0ce1e-af6b-76d1-a39c-84394870f6f3.jsonl', 1727100100, 1727100100, 'cli', 'openai', '/workspace', 'Subagent Luna', 'danger', 'manual'),
('01a0ce22-d3e0-79d0-b8ae-cce0dd14fa20', '$sess_dir/rollout-01a0ce22-d3e0-79d0-b8ae-cce0dd14fa20.jsonl', 1727100200, 1727100200, 'cli', 'openai', '/workspace', 'Subagent Astra', 'danger', 'manual');

CREATE TABLE IF NOT EXISTS thread_spawn_edges (
	parent_thread_id TEXT NOT NULL,
	child_thread_id TEXT NOT NULL PRIMARY KEY,
	status TEXT NOT NULL
);
INSERT OR REPLACE INTO thread_spawn_edges VALUES ('$session_id', '01a0ce1e-af6b-76d1-a39c-84394870f6f3', 'open');
INSERT OR REPLACE INTO thread_spawn_edges VALUES ('$session_id', '01a0ce22-d3e0-79d0-b8ae-cce0dd14fa20', 'open');
SQL
      ;;

    claude)
      local proj_dir="$prof_dir/.claude/projects/-Users-test-aim"
      mkdir -p "$proj_dir"
      local sess_file="$proj_dir/$session_id.jsonl"
      python3 - "$sess_file" "$session_id" << 'PYEOF'
import sys
sess_file, session_id = sys.argv[1], sys.argv[2]
with open(sess_file, "w") as f:
    f.write(f'{{"type":"last-prompt","leafUuid":"leaf-0","sessionId":"{session_id}"}}\n')
    for i in range(50):
        f.write(f'{{"type":"user","message":{{"role":"user","content":"claude-e2e-task prompt {i}"}},"timestamp":"2026-09-23T12:00:00.000Z","sessionId":"{session_id}"}}\n')
        f.write(f'{{"type":"assistant","message":{{"role":"assistant","content":"assistant response {i}"}},"timestamp":"2026-09-23T12:01:00.000Z","sessionId":"{session_id}"}}\n')
PYEOF
      ;;

    agy)
      local brain_dir="$AIM_REAL_HOME/.gemini/antigravity-cli/brain/$session_id/.system_generated/logs"
      mkdir -p "$brain_dir"
      echo '{"content":"<USER_REQUEST>agy-e2e-task implementation request</USER_REQUEST>"}' > "$brain_dir/transcript.jsonl"

      local db="$AIM_REAL_HOME/.gemini/antigravity-cli/conversation_summaries.db"
      sqlite3 "$db" <<SQL
CREATE TABLE IF NOT EXISTS conversation_summaries (conversation_id TEXT PRIMARY KEY, title TEXT, preview TEXT, last_modified_time TEXT);
INSERT OR REPLACE INTO conversation_summaries VALUES ('$session_id', 'agy-e2e-task', 'implementation request', '2026-09-23T12:00:00Z');
SQL
      ;;
  esac
}

# Helper: Setup older/truncated stub in dest profile with a future timestamp (to test size-aware sync)
setup_stale_truncated_stub() {
  local agent="$1"
  local profile="$2"
  local session_id="$3"
  local prof_dir="$AIM_HOME/profiles/$profile"

  case "$agent" in
    codex)
      local sess_dir="$prof_dir/.codex/sessions/2026/09/22"
      mkdir -p "$sess_dir"
      local stub="$sess_dir/rollout-$session_id.jsonl"
      echo '{"event":"turn","turn_index":0,"content":"stale stub"}' > "$stub"
      touch -t 202609281200 "$stub"

      local db="$prof_dir/.codex/state_5.sqlite"
      sqlite3 "$db" <<SQL
CREATE TABLE IF NOT EXISTS _sqlx_migrations (version BIGINT PRIMARY KEY, description TEXT NOT NULL, installed_on TIMESTAMP, success BOOLEAN, checksum BLOB, execution_time BIGINT);
INSERT OR IGNORE INTO _sqlx_migrations VALUES (1, 'init', CURRENT_TIMESTAMP, 1, X'00', 1);
CREATE TABLE IF NOT EXISTS threads (
	id TEXT PRIMARY KEY,
	rollout_path TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	source TEXT NOT NULL,
	model_provider TEXT NOT NULL,
	cwd TEXT NOT NULL,
	title TEXT NOT NULL,
	sandbox_policy TEXT NOT NULL,
	approval_mode TEXT NOT NULL
);
INSERT OR REPLACE INTO threads VALUES ('$session_id', '$stub', 1700000000, 1799999999, 'cli', 'openai', '/workspace', 'stale-stub', 'danger', 'manual');
SQL
      ;;

    claude)
      local proj_dir="$prof_dir/.claude/projects/-Users-test-aim"
      mkdir -p "$proj_dir"
      local stub="$proj_dir/$session_id.jsonl"
      echo '{"type":"user","message":{"role":"user","content":"stale stub"}}' > "$stub"
      touch -t 202609281200 "$stub"
      ;;

    agy)
      # In agy, storage is symlinked/shared across profiles, so no stale stub is needed.
      ;;
  esac
}

# Helper: Verify hydrated session in destination profile
verify_hydration() {
  local agent="$1"
  local profile="$2"
  local session_id="$3"
  local prof_dir="$AIM_HOME/profiles/$profile"

  case "$agent" in
    codex)
      local rollout="$prof_dir/.codex/sessions/2026/09/22/rollout-$session_id.jsonl"
      if [ ! -f "$rollout" ]; then
        echo "FAIL [$agent]: Rollout file missing in destination profile $profile"
        exit 1
      fi
      local size=$(wc -c < "$rollout" | tr -d ' ')
      if [ "$size" -lt 5000 ]; then
        echo "FAIL [$agent]: Rollout file size ($size bytes) in $profile is truncated, expected full rollout"
        exit 1
      fi
      # Verify child subagents
      local child1="$prof_dir/.codex/sessions/2026/09/22/rollout-01a0ce1e-af6b-76d1-a39c-84394870f6f3.jsonl"
      if [ ! -f "$child1" ]; then
        echo "FAIL [$agent]: Subagent rollout missing in destination profile $profile"
        exit 1
      fi
      # Verify SQLite edges
      local db="$prof_dir/.codex/state_5.sqlite"
      local edge_count=$(sqlite3 "$db" "SELECT COUNT(*) FROM thread_spawn_edges WHERE parent_thread_id = '$session_id';")
      if [ "$edge_count" -ne 2 ]; then
        echo "FAIL [$agent]: thread_spawn_edges count in $profile is $edge_count, expected 2"
        exit 1
      fi
      # Verify shell snapshots
      local shell="$prof_dir/.codex/shell_snapshots/snapshot-1.json"
      if [ ! -f "$shell" ]; then
        echo "FAIL [$agent]: Shell snapshot missing in destination profile $profile"
        exit 1
      fi
      ;;

    claude)
      local sess_file="$prof_dir/.claude/projects/-Users-test-aim/$session_id.jsonl"
      if [ ! -f "$sess_file" ]; then
        echo "FAIL [$agent]: Claude session file missing in destination profile $profile"
        exit 1
      fi
      local size=$(wc -c < "$sess_file" | tr -d ' ')
      if [ "$size" -lt 2000 ]; then
        echo "FAIL [$agent]: Claude session file size ($size bytes) is truncated, expected >= 2000"
        exit 1
      fi
      ;;

    agy)
      # Verify agy accesses the brain transcript
      local brain_file="$AIM_REAL_HOME/.gemini/antigravity-cli/brain/$session_id/.system_generated/logs/transcript.jsonl"
      if [ ! -f "$brain_file" ]; then
        echo "FAIL [$agent]: AGY brain transcript missing"
        exit 1
      fi
      ;;
  esac
}

echo ""
echo "========================================================================"
echo "  PHASE 1: Core Launch, Environment Isolation, Flags & Typo Safety      "
echo "  Parameterizing across ALL 4 adapters: agy, codex, claude, gemini       "
echo "========================================================================"

for AGENT in "${ALL_ADAPTERS[@]}"; do
  echo ""
  echo "--- Testing Adapter: [$AGENT] ---"
  PROF="prof-$AGENT"
  LOG="$TEST_DIR/${AGENT}_invoked.log"
  rm -f "$LOG"

  # 1. Test clean 'aim run' with multiple arbitrary forwarded flags
  echo "  [1/4] Testing 'aim run $AGENT $PROF --dangerously-skip-permissions --verbose --custom-flag=test-42'..."
  export AIM_AUTO_CREATE=1
  "$AIM_BIN" run "$AGENT" "$PROF" --dangerously-skip-permissions --verbose --custom-flag=test-42

  if [ ! -f "$LOG" ]; then
    echo "FAIL [$AGENT]: Mock binary was not invoked"
    exit 1
  fi

  if ! grep -q "ARGS: .*--dangerously-skip-permissions" "$LOG"; then
    echo "FAIL [$AGENT]: Flag '--dangerously-skip-permissions' was swallowed or dropped"
    cat "$LOG"
    exit 1
  fi
  if ! grep -q "ARGS: .*--verbose" "$LOG"; then
    echo "FAIL [$AGENT]: Flag '--verbose' was swallowed or dropped"
    cat "$LOG"
    exit 1
  fi
  if ! grep -q "ARGS: .*--custom-flag=test-42" "$LOG"; then
    echo "FAIL [$AGENT]: Arbitrary flag '--custom-flag=test-42' was swallowed or dropped"
    cat "$LOG"
    exit 1
  fi

  # Verify environment isolation
  EXPECTED_HOME="$AIM_HOME/profiles/$PROF"
  if ! grep -q "HOME: $EXPECTED_HOME" "$LOG"; then
    echo "FAIL [$AGENT]: HOME not isolated to profile dir (expected $EXPECTED_HOME)"
    cat "$LOG"
    exit 1
  fi
  if ! grep -q "AIM_AGENT: $AGENT" "$LOG"; then
    echo "FAIL [$AGENT]: AIM_AGENT environment variable not set to $AGENT"
    cat "$LOG"
    exit 1
  fi
  if ! grep -q "AIM_PROFILE: $PROF" "$LOG"; then
    echo "FAIL [$AGENT]: AIM_PROFILE environment variable not set to $PROF"
    cat "$LOG"
    exit 1
  fi

  # Verify dotfiles were bridged into profile
  if [ ! -f "$EXPECTED_HOME/.gitconfig" ]; then
    echo "FAIL [$AGENT]: .gitconfig was not provisioned in profile directory"
    exit 1
  fi
  echo "  ✔ Flag forwarding, environment isolation, and dotfile bridging verified for $AGENT"

  # 2. Test typo guard in non-interactive / automated script mode
  echo "  [2/4] Testing non-interactive typo guard for 'aim run $AGENT typo--dangerously-skip-permissions'..."
  NON_INT_OUT=$("$AIM_BIN" run "$AGENT" typo--dangerously-skip-permissions 2>&1 || true)
  if ! echo "$NON_INT_OUT" | grep -q 'refusing to auto-create profile.*containing "--"'; then
    echo "FAIL [$AGENT]: Non-interactive run did not reject dashed profile typo"
    echo "$NON_INT_OUT"
    exit 1
  fi
  if ! echo "$NON_INT_OUT" | grep -q 'did you mean "typo" with flag "--dangerously-skip-permissions"'; then
    echo "FAIL [$AGENT]: Non-interactive typo error missing suggestion"
    echo "$NON_INT_OUT"
    exit 1
  fi
  echo "  ✔ Non-interactive typo guard blocked invalid profile creation"

  # 3. Test interactive typo prompt via PTY
  echo "  [3/4] Testing interactive typo prompt with Lipgloss formatting..."
  PTY_OUT=$(python3 -c '
import pty, os, subprocess

master, slave = pty.openpty()
env = os.environ.copy()
env.pop("AIM_AUTO_CREATE", None)
proc = subprocess.Popen(["'"$AIM_BIN"'", "run", "'"$AGENT"'", "typo--dangerously-skip-permissions"], stdin=slave, stdout=slave, stderr=slave, close_fds=True, env=env)
os.close(slave)
sent = False
out = b""
while True:
    try:
        data = os.read(master, 1024)
        if not data:
            break
        out += data
        if not sent and b"[y/N]:" in out:
            os.write(master, b"n\n")
            sent = True
    except OSError:
        break
os.close(master)
proc.wait()
print(out.decode("utf-8", errors="replace"))
')

  CLEAN_PTY=$(echo "$PTY_OUT" | sed -E $'s/\x1b\\[[0-9;]*m//g')
  if ! echo "$CLEAN_PTY" | grep -q 'Warning: profile name.*contains "--"'; then
    echo "FAIL [$AGENT]: Interactive PTY did not show dashed typo warning"
    echo "$PTY_OUT"
    exit 1
  fi
  if ! echo "$CLEAN_PTY" | grep -q 'Profile creation aborted'; then
    echo "FAIL [$AGENT]: Interactive PTY abort failed"
    echo "$PTY_OUT"
    exit 1
  fi
  echo "  ✔ Interactive typo prompt warned and aborted cleanly"

  # 4. Test aim doctor
  echo "  [4/4] Testing 'aim doctor $AGENT'..."
  DOCTOR_OUT=$("$AIM_BIN" doctor "$AGENT")
  if ! echo "$DOCTOR_OUT" | grep -q "AIM Doctor Diagnostics"; then
    echo "FAIL [$AGENT]: 'aim doctor' output unexpected"
    echo "$DOCTOR_OUT"
    exit 1
  fi
  echo "  ✔ Doctor diagnostics passed for $AGENT"
done

echo ""
echo "========================================================================"
echo "  PHASE 2: Session Lifecycle, Discovery, Cross-Profile Resumption,       "
echo "           Size-Aware Sync, and Forking                                 "
echo "  Parameterizing across SESSION adapters: agy, codex, claude             "
echo "========================================================================"

SESSION_ADAPTERS=("codex" "claude" "agy")

for AGENT in "${SESSION_ADAPTERS[@]}"; do
  echo ""
  echo "--- Testing Session Lifecycle: [$AGENT] ---"
  SRC_PROF="work"
  DST_PROF="office"
  LOG="$TEST_DIR/${AGENT}_invoked.log"
  rm -f "$LOG"

  # Generate unique session ID for this agent run
  SESS_ID="01a0-test-$AGENT-$(python3 -c 'import uuid; print(str(uuid.uuid4())[:8])')"
  if [ "$AGENT" = "claude" ] || [ "$AGENT" = "agy" ]; then
    SESS_ID=$(python3 -c 'import uuid; print(str(uuid.uuid4()))')
  elif [ "$AGENT" = "codex" ]; then
    SESS_ID="01a0c7e1-$(python3 -c 'import uuid; print(str(uuid.uuid4())[9:])')"
  fi

  # 1. Create complete multi-turn session in 'work'
  echo "  [1/4] Setting up rich multi-turn session ($SESS_ID) in '$SRC_PROF'..."
  setup_test_session "$AGENT" "$SRC_PROF" "$SESS_ID"

  # 2. Test session discovery via 'aim sessions'
  echo "  [2/4] Testing 'aim sessions --agent $AGENT' discovery..."
  SESSIONS_OUT=$("$AIM_BIN" sessions --agent "$AGENT")
  SHORT_ID="${SESS_ID:0:8}"
  if ! echo "$SESSIONS_OUT" | grep -q "$SHORT_ID"; then
    echo "FAIL [$AGENT]: 'aim sessions' did not discover session $SHORT_ID"
    echo "$SESSIONS_OUT"
    exit 1
  fi
  echo "  ✔ Session discovered in session list"

  # 3. Simulate existing older/truncated stub in 'office' with newer timestamp to test size-aware sync
  echo "  [3/4] Testing size-aware cross-profile resumption & flag forwarding..."
  setup_stale_truncated_stub "$AGENT" "$DST_PROF" "$SESS_ID"

  # Resume into destination profile with arbitrary flags
  "$AIM_BIN" resume "$AGENT" "$DST_PROF" "$SHORT_ID" --dangerously-skip-permissions --verbose

  # Verify agent was invoked with flags
  if ! grep -q "ARGS: .*--dangerously-skip-permissions" "$LOG"; then
    echo "FAIL [$AGENT]: Resume did not forward '--dangerously-skip-permissions'"
    cat "$LOG"
    exit 1
  fi
  if ! grep -q "ARGS: .*--verbose" "$LOG"; then
    echo "FAIL [$AGENT]: Resume did not forward '--verbose'"
    cat "$LOG"
    exit 1
  fi

  # Verify resume command received session identifier
  case "$AGENT" in
    codex)
      if ! grep -q "ARGS: resume $SESS_ID" "$LOG"; then
        echo "FAIL [$AGENT]: Codex resume invocation did not include 'resume $SESS_ID'"
        cat "$LOG"
        exit 1
      fi
      ;;
    claude)
      if ! grep -q "ARGS: --resume $SESS_ID" "$LOG"; then
        echo "FAIL [$AGENT]: Claude resume invocation did not include '--resume $SESS_ID'"
        cat "$LOG"
        exit 1
      fi
      ;;
    agy)
      if ! grep -q "ARGS: --conversation=$SESS_ID" "$LOG"; then
        echo "FAIL [$AGENT]: AGY resume invocation did not include '--conversation=$SESS_ID'"
        cat "$LOG"
        exit 1
      fi
      ;;
  esac

  # Verify destination profile received full content
  verify_hydration "$AGENT" "$DST_PROF" "$SESS_ID"
  echo "  ✔ Full session hydrated into destination profile '$DST_PROF' with forwarded flags"

  # 4. Test session forking (--fork) if supported by adapter
  if adapter_supports_fork "$AGENT"; then
    echo "  [4/4] Testing session forking ('aim resume $AGENT $DST_PROF $SHORT_ID --fork')..."
    rm -f "$LOG"
    "$AIM_BIN" resume "$AGENT" "$DST_PROF" "$SHORT_ID" --fork --dangerously-skip-permissions

    if ! grep -q "ARGS: .*--dangerously-skip-permissions" "$LOG"; then
      echo "FAIL [$AGENT]: Fork resume did not forward flags"
      cat "$LOG"
      exit 1
    fi
    echo "  ✔ Forked session generated and launched cleanly"
  else
    echo "  [4/4] Forking not applicable for $AGENT (skipping)"
  fi
done

echo ""
echo "========================================================================"
echo "  PHASE 3: Non-Session Adapter Rejection Handling                        "
echo "  Verifying graceful error when resuming non-session agents (gemini)     "
echo "========================================================================"

NON_SESSION_ADAPTERS=("gemini")

for AGENT in "${NON_SESSION_ADAPTERS[@]}"; do
  echo "--- Testing Non-Session Rejection: [$AGENT] ---"
  NON_SESS_OUT=$("$AIM_BIN" resume "$AGENT" office dummy-sess-id 2>&1 || true)
  if ! echo "$NON_SESS_OUT" | grep -qi 'session.*not'; then
    echo "FAIL [$AGENT]: Expected session not supported/found error for non-session agent"
    echo "$NON_SESS_OUT"
    exit 1
  fi
  echo "✔ Correctly rejected resumption for non-session adapter '$AGENT'"
done

echo ""
echo "========================================================================"
echo "  PHASE 4: Session-Scoped MCP Server Merge                               "
echo "  (global merge, keep, promote, mcp_global:false, background recovery)   "
echo "  Parameterizing across ALL adapters; MCP: claude, codex, agy            "
echo "========================================================================"

# Helper: Check if adapter merges host MCP servers per session (agents.MCPProvider)
adapter_has_mcp() {
  case "$1" in
    claude|codex|agy) return 0 ;;
    *) return 1 ;;
  esac
}

# Helper: Host-side MCP config file for adapter (under AIM_REAL_HOME)
mcp_host_file() {
  case "$1" in
    claude) echo "$AIM_REAL_HOME/.claude.json" ;;
    codex)  echo "$AIM_REAL_HOME/.codex/config.toml" ;;
    agy)    echo "$AIM_REAL_HOME/.gemini/config/mcp_config.json" ;;
  esac
}

# Helper: Profile-side MCP config file, relative to the profile dir ($HOME inside the agent)
mcp_profile_rel() {
  case "$1" in
    claude) echo ".claude/.claude.json" ;;
    codex)  echo ".codex/config.toml" ;;
    agy)    echo ".gemini/config/mcp_config.json" ;;
  esac
}

# Helper: File format and top-level key of the MCP server map
mcp_format() { case "$1" in codex) echo toml ;; *) echo json ;; esac; }
mcp_key()    { case "$1" in codex) echo mcp_servers ;; *) echo mcpServers ;; esac; }

# Helper: Native args that start a background session (merge, run, no exit step)
mcp_bg_args() {
  case "$1" in
    claude) echo "--bg" ;;
    codex)  echo "app-server" ;;
    agy)    echo "remote-control" ;;
  esac
}

# Helper: Write an MCP config holding the named servers (server -> command "true", plus args for 'shared')
mcp_write() {
  local agent="$1" file="$2"; shift 2
  mkdir -p "$(dirname "$file")"
  python3 - "$(mcp_format "$agent")" "$(mcp_key "$agent")" "$file" "$@" << 'PYEOF'
import json, sys
fmt, key, path, names = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4:]
servers = {n: ({"command": "true", "args": ["x"]} if n == "shared" else {"command": "true"}) for n in names}
if fmt == "json":
    json.dump({"other": 1, key: servers}, open(path, "w"), indent=2)
else:
    with open(path, "w") as f:
        f.write('model = "e2e"\n')
        for n, v in servers.items():
            f.write(f'\n[{key}.{n}]\ncommand = "{v["command"]}"\n')
            if "args" in v:
                f.write('args = ["x"]\n')
PYEOF
}

# Helper: Print the server names in an MCP config, space-separated and sorted
mcp_names() {
  local agent="$1" file="$2"
  [ -f "$file" ] || { echo ""; return; }
  python3 - "$(mcp_format "$agent")" "$(mcp_key "$agent")" "$file" << 'PYEOF'
import json, sys, tomllib
fmt, key, path = sys.argv[1], sys.argv[2], sys.argv[3]
data = json.load(open(path)) if fmt == "json" else tomllib.load(open(path, "rb"))
print(" ".join(sorted((data.get(key) or {}).keys())))
PYEOF
}

# Helper: Assert an MCP config holds exactly the given servers
assert_mcp_names() {
  local agent="$1" file="$2" what="$3"; shift 3
  local want got
  want=$(printf '%s\n' "$@" | sort | tr '\n' ' ' | sed 's/ $//')
  got=$(mcp_names "$agent" "$file")
  if [ "$got" != "$want" ]; then
    echo "FAIL [$agent]: $what has servers [$got], expected [$want]"
    [ -f "$file" ] && cat "$file"
    exit 1
  fi
}

# Helper: Set or clear profiles.<p>.mcp_global in aim's config.json
set_mcp_global() {
  python3 - "$AIM_HOME/config.json" "$1" "$2" << 'PYEOF'
import json, sys
path, prof, val = sys.argv[1], sys.argv[2], sys.argv[3]
cfg = json.load(open(path))
p = cfg.setdefault("profiles", {}).setdefault(prof, {})
if val == "unset":
    p.pop("mcp_global", None)
else:
    p["mcp_global"] = (val == "true")
json.dump(cfg, open(path, "w"), indent=2)
PYEOF
}

# Helper: Set profiles.<p>.args in aim's config.json to the remaining arguments, or clear it with "unset"
set_profile_args() {
  python3 - "$AIM_HOME/config.json" "$@" << 'PYEOF'
import json, sys
path, prof, args = sys.argv[1], sys.argv[2], sys.argv[3:]
cfg = json.load(open(path))
p = cfg.setdefault("profiles", {}).setdefault(prof, {})
if args == ["unset"]:
    p.pop("args", None)
else:
    p["args"] = args
json.dump(cfg, open(path, "w"), indent=2)
PYEOF
}

# Helper: Run a command on a PTY, answer once when EXPECT appears, print output
pty_run() {
  local expect="$1" answer="$2"; shift 2
  python3 - "$expect" "$answer" "$@" << 'PYEOF'
import os, pty, subprocess, sys
expect, answer, cmd = sys.argv[1].encode(), sys.argv[2].encode(), sys.argv[3:]
master, slave = pty.openpty()
proc = subprocess.Popen(cmd, stdin=slave, stdout=slave, stderr=slave, close_fds=True)
os.close(slave)
sent, out = False, b""
while True:
    try:
        data = os.read(master, 1024)
    except OSError:
        break
    if not data:
        break
    out += data
    if not sent and expect in out:
        os.write(master, answer + b"\n")
        sent = True
os.close(master)
proc.wait()
sys.stdout.write(out.decode("utf-8", errors="replace"))
sys.exit(proc.returncode)
PYEOF
}

# Hook the mock agent runs mid-session (E2E_HOOK): snapshot the profile's MCP
# config, then add a server natively the way `<agent> mcp add` would.
cat << 'HOOK' > "$TEST_DIR/mcp_hook.sh"
#!/bin/sh
file="$HOME/$E2E_MCP_REL"
cp "$file" "$E2E_SNAP" 2>/dev/null || : > "$E2E_SNAP"
[ -n "${E2E_MCP_ADD:-}" ] || exit 0
if [ "$E2E_MCP_FORMAT" = "toml" ]; then
  printf '\n[%s.%s]\ncommand = "true"\n' "$E2E_MCP_KEY" "$E2E_MCP_ADD" >> "$file"
else
  python3 - "$file" "$E2E_MCP_KEY" "$E2E_MCP_ADD" << 'PYEOF'
import json, sys
path, key, name = sys.argv[1], sys.argv[2], sys.argv[3]
data = json.load(open(path))
data.setdefault(key, {})[name] = {"command": "true"}
json.dump(data, open(path, "w"), indent=2)
PYEOF
fi
HOOK

export E2E_HOOK="$TEST_DIR/mcp_hook.sh"

for AGENT in "${ALL_ADAPTERS[@]}"; do
  echo ""
  echo "--- Testing MCP Merge: [$AGENT] ---"
  PROF="mcp-$AGENT"
  PROF_DIR="$AIM_HOME/profiles/$PROF"

  if ! adapter_has_mcp "$AGENT"; then
    export AIM_AUTO_CREATE=1
    "$AIM_BIN" run "$AGENT" "$PROF" < /dev/null
    if [ -f "$AIM_HOME/profile-merge/$PROF.json" ]; then
      echo "FAIL [$AGENT]: adapter without MCP servers left merge state at profile-merge/$PROF.json"
      exit 1
    fi
    echo "  ✔ No MCP merge for $AGENT (no MCPProvider), launch unaffected"
    continue
  fi

  HOST_FILE=$(mcp_host_file "$AGENT")
  PROF_REL=$(mcp_profile_rel "$AGENT")
  PROF_FILE="$PROF_DIR/$PROF_REL"
  SNAP="$TEST_DIR/${AGENT}_mcp_during"
  export E2E_MCP_REL="$PROF_REL" E2E_MCP_KEY="$(mcp_key "$AGENT")" E2E_MCP_FORMAT="$(mcp_format "$AGENT")" E2E_SNAP="$SNAP"

  # Host: hostsrv + shared. Profile: own + an identical copy-once leftover of shared.
  mcp_write "$AGENT" "$HOST_FILE" hostsrv shared
  mcp_write "$AGENT" "$PROF_FILE" own shared
  cp "$HOST_FILE" "$TEST_DIR/${AGENT}_host_before"

  # 1. Global merge visible during the session; a native add is kept in the profile on exit
  echo "  [1/6] Testing global merge, one-time migration and keep-on-exit (non-interactive)..."
  export AIM_AUTO_CREATE=1
  export E2E_MCP_ADD=newsrv
  RUN_OUT=$("$AIM_BIN" run "$AGENT" "$PROF" < /dev/null 2>&1)
  assert_mcp_names "$AGENT" "$SNAP" "profile config during session" own hostsrv shared
  assert_mcp_names "$AGENT" "$PROF_FILE" "profile config at rest" own newsrv
  if ! echo "$RUN_OUT" | grep -q "1 change(s) kept in $PROF ($AGENT)"; then
    echo "FAIL [$AGENT]: exit did not report the kept change"
    echo "$RUN_OUT"
    exit 1
  fi
  if ! ls "$PROF_FILE".aim-backup-* > /dev/null 2>&1; then
    echo "FAIL [$AGENT]: migration removed the identical 'shared' copy without a backup"
    exit 1
  fi
  if ! cmp -s "$HOST_FILE" "$TEST_DIR/${AGENT}_host_before"; then
    echo "FAIL [$AGENT]: host config changed by a keep-only session"
    exit 1
  fi
  echo "  ✔ Host servers merged for the session only; 'newsrv' kept, 'shared' leftover migrated with backup"

  # 2. Promote: answer [p] at the exit prompt on a PTY; the host gains the server
  echo "  [2/6] Testing promote to host via exit prompt (PTY)..."
  export E2E_MCP_ADD=promoted
  PTY_OUT=$(pty_run "(default: k)" "p" "$AIM_BIN" run "$AGENT" "$PROF")
  CLEAN_PTY=$(echo "$PTY_OUT" | sed -E $'s/\x1b\\[[0-9;]*m//g')
  if ! echo "$CLEAN_PTY" | grep -q "Session in $PROF ($AGENT) changed:"; then
    echo "FAIL [$AGENT]: exit prompt not shown on a TTY"
    echo "$PTY_OUT"
    exit 1
  fi
  assert_mcp_names "$AGENT" "$HOST_FILE" "host config after promote" hostsrv shared promoted
  assert_mcp_names "$AGENT" "$PROF_FILE" "profile config after promote" own newsrv
  echo "  ✔ 'promoted' written to the host and stripped from the profile"

  # 3. mcp_global: false — the profile runs on its own servers only
  echo "  [3/6] Testing profiles.$PROF.mcp_global = false..."
  set_mcp_global "$PROF" false
  unset E2E_MCP_ADD
  "$AIM_BIN" run "$AGENT" "$PROF" < /dev/null
  assert_mcp_names "$AGENT" "$SNAP" "profile config during session (mcp_global=false)" own newsrv
  set_mcp_global "$PROF" unset
  echo "  ✔ Host servers not merged when mcp_global is false"

  # 4. Background launch: merge, run, no exit step — host items stay until the next launch
  echo "  [4/6] Testing background launch ('aim run $AGENT $PROF $(mcp_bg_args "$AGENT")')..."
  export E2E_MCP_ADD=bgsrv
  "$AIM_BIN" run "$AGENT" "$PROF" $(mcp_bg_args "$AGENT") < /dev/null
  unset E2E_MCP_ADD
  assert_mcp_names "$AGENT" "$PROF_FILE" "profile config after background launch" own newsrv bgsrv hostsrv shared promoted
  echo "  ✔ Background launch left the merged config in place"

  # 5. Background via profiles.<p>.args with no CLI args: recovers step 4, merges, no exit step
  echo "  [5/6] Testing background launch from profiles.$PROF.args = [$(mcp_bg_args "$AGENT")]..."
  set_profile_args "$PROF" $(mcp_bg_args "$AGENT")
  export E2E_MCP_ADD=profbgsrv
  RUN_OUT=$("$AIM_BIN" run "$AGENT" "$PROF" < /dev/null 2>&1)
  unset E2E_MCP_ADD
  set_profile_args "$PROF" unset
  if ! echo "$RUN_OUT" | grep -q "1 change(s) kept in $PROF ($AGENT) from a session that did not exit through aim"; then
    echo "FAIL [$AGENT]: launch from profile args did not recover the previous background session"
    echo "$RUN_OUT"
    exit 1
  fi
  assert_mcp_names "$AGENT" "$PROF_FILE" "profile config after background launch from profile args" own newsrv bgsrv profbgsrv hostsrv shared promoted
  echo "  ✔ Profile args made the launch a background one: host items left in place"

  # 6. The next foreground launch recovers: keeps the background session's change, strips host items
  echo "  [6/6] Testing recovery on the next launch..."
  RUN_OUT=$("$AIM_BIN" run "$AGENT" "$PROF" < /dev/null 2>&1)
  if ! echo "$RUN_OUT" | grep -q "1 change(s) kept in $PROF ($AGENT) from a session that did not exit through aim"; then
    echo "FAIL [$AGENT]: recovery did not report the background session's change"
    echo "$RUN_OUT"
    exit 1
  fi
  assert_mcp_names "$AGENT" "$SNAP" "profile config during recovered session" own newsrv bgsrv profbgsrv hostsrv shared promoted
  assert_mcp_names "$AGENT" "$PROF_FILE" "profile config at rest after recovery" own newsrv bgsrv profbgsrv
  assert_mcp_names "$AGENT" "$HOST_FILE" "host config after recovery" hostsrv shared promoted
  echo "  ✔ Recovered: 'bgsrv' and 'profbgsrv' kept, host items stripped, host untouched"
done

unset E2E_HOOK E2E_MCP_ADD E2E_MCP_REL E2E_MCP_KEY E2E_MCP_FORMAT E2E_SNAP

echo ""
echo "========================================================================"
echo "  ALL PARAMETERIZED E2E TESTS PASSED ACROSS ALL 4 ADAPTERS!              "
echo "========================================================================"
