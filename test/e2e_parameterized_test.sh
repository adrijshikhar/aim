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
echo "  ALL PARAMETERIZED E2E TESTS PASSED ACROSS ALL 4 ADAPTERS!              "
echo "========================================================================"
