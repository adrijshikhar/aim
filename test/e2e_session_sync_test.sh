#!/usr/bin/env bash
set -euo pipefail

echo "========================================================"
echo "  E2E Test: Session Lifecycle, Sync & Flags Execution  "
echo "========================================================"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AIM_BIN="$REPO_ROOT/aim"
(cd "$REPO_ROOT" && go build -o "$AIM_BIN" ./cmd/aim)

TEST_DIR=$(mktemp -d)
MOCK_BIN=$(mktemp -d)
trap 'rm -rf "$TEST_DIR" "$MOCK_BIN"' EXIT

export AIM_HOME="$TEST_DIR/aim_home"
export AIM_REAL_HOME="$TEST_DIR/real_home"
export AIM_AUTO_CREATE=1
export PATH="$MOCK_BIN:$PATH"

mkdir -p "$AIM_HOME/profiles/work/.codex/sessions/2026/09/22"
mkdir -p "$AIM_HOME/profiles/work/.codex/shell_snapshots"
mkdir -p "$AIM_HOME/profiles/office/.codex/sessions/2026/09/22"
mkdir -p "$AIM_HOME/profiles/office/.codex/shell_snapshots"
mkdir -p "$AIM_REAL_HOME"

# Mock codex binary that records invocations
cat << MOCK > "$MOCK_BIN/codex"
#!/bin/sh
if [ "\$1" = "--version" ]; then
  echo "codex-cli 0.154.0"
  exit 0
fi
echo "CODEX_INVOCATION: \$*" > "$TEST_DIR/codex_invoked.log"
MOCK
chmod +x "$MOCK_BIN/codex"

PARENT_ID="01a0c7e1-e34c-7ca2-be3f-df5db5e94752"
CHILD1_ID="01a0ce1e-af6b-76d1-a39c-84394870f6f3"
CHILD2_ID="01a0ce22-d3e0-79d0-b8ae-cce0dd14fa20"

# 1. In 'work': Create complete session with large rollout (30KB), 2 subagents, shell snapshot
WORK_PARENT_ROLLOUT="$AIM_HOME/profiles/work/.codex/sessions/2026/09/22/rollout-$PARENT_ID.jsonl"
python3 -c '
with open("'$WORK_PARENT_ROLLOUT'", "w") as f:
    for i in range(100):
        f.write(f"{{\"event\":\"turn\",\"turn_index\":{i},\"content\":\"turn content {i} with extensive tooling work\"}}\n")
'
WORK_CHILD1_ROLLOUT="$AIM_HOME/profiles/work/.codex/sessions/2026/09/22/rollout-$CHILD1_ID.jsonl"
echo '{"event":"subagent_luna_init"}' > "$WORK_CHILD1_ROLLOUT"

WORK_CHILD2_ROLLOUT="$AIM_HOME/profiles/work/.codex/sessions/2026/09/22/rollout-$CHILD2_ID.jsonl"
echo '{"event":"subagent_astra_init"}' > "$WORK_CHILD2_ROLLOUT"

echo '{"shell_snapshot":"active_term"}' > "$AIM_HOME/profiles/work/.codex/shell_snapshots/snapshot-1.json"

WORK_STATE_DB="$AIM_HOME/profiles/work/.codex/state_5.sqlite"
sqlite3 "$WORK_STATE_DB" <<SQL
CREATE TABLE _sqlx_migrations (version BIGINT PRIMARY KEY, description TEXT NOT NULL, installed_on TIMESTAMP, success BOOLEAN, checksum BLOB, execution_time BIGINT);
INSERT INTO _sqlx_migrations VALUES (1, 'init', CURRENT_TIMESTAMP, 1, X'00', 1);

CREATE TABLE threads (
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
INSERT INTO threads VALUES
('$PARENT_ID', '$WORK_PARENT_ROLLOUT', 1727100000, 1727100000, 'cli', 'openai', '/workspace', 'DSL Test Connection Full', 'danger', 'manual'),
('$CHILD1_ID', '$WORK_CHILD1_ROLLOUT', 1727100100, 1727100100, 'cli', 'openai', '/workspace', 'Subagent Luna', 'danger', 'manual'),
('$CHILD2_ID', '$WORK_CHILD2_ROLLOUT', 1727100200, 1727100200, 'cli', 'openai', '/workspace', 'Subagent Astra', 'danger', 'manual');

CREATE TABLE thread_spawn_edges (
	parent_thread_id TEXT NOT NULL,
	child_thread_id TEXT NOT NULL PRIMARY KEY,
	status TEXT NOT NULL
);
INSERT INTO thread_spawn_edges VALUES
('$PARENT_ID', '$CHILD1_ID', 'open'),
('$PARENT_ID', '$CHILD2_ID', 'open');
SQL

# 2. In 'office': Create a truncated/stale session (only 1KB), BUT touched today (updated_at = NOW)
OFFICE_PARENT_ROLLOUT="$AIM_HOME/profiles/office/.codex/sessions/2026/09/22/rollout-$PARENT_ID.jsonl"
echo '{"event":"turn","turn_index":1,"content":"truncated old snapshot"}' > "$OFFICE_PARENT_ROLLOUT"

NOW=$(date +%s)
OFFICE_STATE_DB="$AIM_HOME/profiles/office/.codex/state_5.sqlite"
sqlite3 "$OFFICE_STATE_DB" <<SQL
CREATE TABLE _sqlx_migrations (version BIGINT PRIMARY KEY, description TEXT NOT NULL, installed_on TIMESTAMP, success BOOLEAN, checksum BLOB, execution_time BIGINT);
INSERT INTO _sqlx_migrations VALUES (1, 'init', CURRENT_TIMESTAMP, 1, X'00', 1);

CREATE TABLE threads (
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
INSERT INTO threads VALUES
('$PARENT_ID', '$OFFICE_PARENT_ROLLOUT', 1727100000, $NOW, 'cli', 'openai', '/workspace', 'DSL Test Connection Truncated', 'danger', 'manual');

CREATE TABLE thread_spawn_edges (
	parent_thread_id TEXT NOT NULL,
	child_thread_id TEXT NOT NULL PRIMARY KEY,
	status TEXT NOT NULL
);
SQL

echo "==> Setup complete: 'work' has 100 turns (complete), 'office' has 1 turn (stale touch timestamp $NOW)"

# 3. Execute resume into 'office' with extra flags
echo "==> Testing 'aim resume codex office 01a0c7e1 --dangerously-skip-permissions --verbose'..."
RESUME_OUTPUT=$("$AIM_BIN" resume codex office 01a0c7e1 --dangerously-skip-permissions --verbose)
echo "$RESUME_OUTPUT"

# 4. Verify codex was launched with resume and passed flags
if ! grep -q "CODEX_INVOCATION: resume $PARENT_ID --dangerously-skip-permissions --verbose" "$TEST_DIR/codex_invoked.log"; then
  echo "FAIL: codex was not invoked with expected resume args and flags"
  cat "$TEST_DIR/codex_invoked.log"
  exit 1
fi
echo "✔ Codex invoked with exact session ID and forwarded flags"

# 5. Verify the larger rollout was synced into 'office'
WORK_SIZE=$(wc -c < "$WORK_PARENT_ROLLOUT" | tr -d ' ')
OFFICE_SIZE=$(wc -c < "$OFFICE_PARENT_ROLLOUT" | tr -d ' ')
if [ "$OFFICE_SIZE" -ne "$WORK_SIZE" ]; then
  echo "FAIL: office rollout was not updated with the complete 100-turn session from work (office: $OFFICE_SIZE, work: $WORK_SIZE)"
  exit 1
fi
echo "✔ Office rollout updated with full content ($OFFICE_SIZE bytes)"

# 6. Verify child subagent rollouts were copied to 'office'
OFFICE_CHILD1="$AIM_HOME/profiles/office/.codex/sessions/2026/09/22/rollout-$CHILD1_ID.jsonl"
OFFICE_CHILD2="$AIM_HOME/profiles/office/.codex/sessions/2026/09/22/rollout-$CHILD2_ID.jsonl"
if [ ! -f "$OFFICE_CHILD1" ] || [ ! -f "$OFFICE_CHILD2" ]; then
  echo "FAIL: child subagent rollouts were not copied to office profile"
  exit 1
fi
echo "✔ Child subagents rollouts successfully hydrated into office profile"

# 7. Verify thread_spawn_edges in office state_5.sqlite
EDGE_COUNT=$(sqlite3 "$OFFICE_STATE_DB" "SELECT COUNT(*) FROM thread_spawn_edges WHERE parent_thread_id = '$PARENT_ID';")
if [ "$EDGE_COUNT" -ne 2 ]; then
  echo "FAIL: thread_spawn_edges count in office is $EDGE_COUNT, expected 2"
  exit 1
fi
echo "✔ Subagent spawn edges (count=$EDGE_COUNT) verified in destination SQLite"

# 8. Verify shell snapshot was copied to 'office'
OFFICE_SHELL="$AIM_HOME/profiles/office/.codex/shell_snapshots/snapshot-1.json"
if [ ! -f "$OFFICE_SHELL" ]; then
  echo "FAIL: shell snapshot was not copied to office profile"
  exit 1
fi
echo "✔ Shell snapshots verified in destination profile"

# 9. Test passing flags through 'aim run'
echo "==> Testing 'aim run agy office --dangerously-skip-permissions'..."
cat << MOCK > "$MOCK_BIN/agy"
#!/bin/sh
echo "AGY_INVOCATION: \$*" > "$TEST_DIR/agy_invoked.log"
MOCK
chmod +x "$MOCK_BIN/agy"

"$AIM_BIN" run agy office --dangerously-skip-permissions
if ! grep -q "AGY_INVOCATION: --dangerously-skip-permissions" "$TEST_DIR/agy_invoked.log"; then
  echo "FAIL: agy was not invoked with expected forwarded flags"
  cat "$TEST_DIR/agy_invoked.log"
  exit 1
fi
echo "✔ 'aim run' cleanly forwarded flags to agent"

# 10. Test typo warning for dashed profile name
echo "==> Testing typo detection for 'aim run agy rs--dangerously-skip-permissions'..."

# 10a. Non-interactive / auto-create guard
NON_INTERACTIVE_OUTPUT=$("$AIM_BIN" run agy rs--dangerously-skip-permissions 2>&1 || true)
if ! echo "$NON_INTERACTIVE_OUTPUT" | grep -q 'refusing to auto-create profile.*containing "--"'; then
  echo "FAIL: expected refusal to auto-create profile with '--'"
  echo "$NON_INTERACTIVE_OUTPUT"
  exit 1
fi
if ! echo "$NON_INTERACTIVE_OUTPUT" | grep -q 'did you mean "rs" with flag "--dangerously-skip-permissions"'; then
  echo "FAIL: expected suggestion in non-interactive dash typo output"
  echo "$NON_INTERACTIVE_OUTPUT"
  exit 1
fi
echo "✔ Non-interactive auto-create guard safely rejected profile containing '--'"

# 10b. Interactive PTY prompt with lipgloss warning
INTERACTIVE_OUTPUT=$(python3 -c '
import pty, os, subprocess

master, slave = pty.openpty()
env = os.environ.copy()
env.pop("AIM_AUTO_CREATE", None)
proc = subprocess.Popen(["'"$AIM_BIN"'", "run", "agy", "rs--dangerously-skip-permissions"], stdin=slave, stdout=slave, stderr=slave, close_fds=True, env=env)
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

CLEAN_OUTPUT=$(echo "$INTERACTIVE_OUTPUT" | sed -E $'s/\x1b\\[[0-9;]*m//g')
if ! echo "$CLEAN_OUTPUT" | grep -q 'Warning: profile name.*contains "--"'; then
  echo "FAIL: expected dash typo warning in interactive prompt"
  echo "$INTERACTIVE_OUTPUT"
  exit 1
fi
if ! echo "$CLEAN_OUTPUT" | grep -q 'Profile creation aborted'; then
  echo "FAIL: expected abort confirmation in interactive prompt"
  echo "$INTERACTIVE_OUTPUT"
  exit 1
fi
echo "✔ Interactive prompt warned with suggestion and cleanly handled abort"

echo ""
echo "========================================================"
echo "  All Live E2E Session Lifecycle Tests PASSED!        "
echo "========================================================"
