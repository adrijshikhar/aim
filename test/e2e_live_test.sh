#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# E2E Live Test Suite for AIM
# Tests live binary execution, PTY interaction, Bubble Tea TUI rendering,
# latency benchmarks, windowing, drawer previews, CLI tables, JSON,
# transparent MCP interception, and command guardrails.
# ==============================================================================

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export AIM_BIN="$REPO_ROOT/aim"

echo "========================================================================"
echo "  AIM LIVE E2E TEST SUITE                                               "
echo "========================================================================"

# 1. Compile fresh binary
echo "==> Compiling fresh binary..."
(cd "$REPO_ROOT" && go build -o "$AIM_BIN" ./cmd/aim)
[ -x "$AIM_BIN" ] || { echo "FAIL: Binary failed to compile"; exit 1; }

# 2. Setup isolated sandbox environment
TEST_SANDBOX=$(mktemp -d)
MOCK_BIN=$(mktemp -d)
trap 'rm -rf "$TEST_SANDBOX" "$MOCK_BIN"' EXIT

export AIM_HOME="$TEST_SANDBOX/aim_home"
export AIM_REAL_HOME="$TEST_SANDBOX/real_home"
export AIM_AUTO_CREATE=1
export PATH="$MOCK_BIN:$PATH"

mkdir -p "$AIM_HOME/profiles/work" "$AIM_HOME/profiles/office" "$AIM_HOME/profiles/staging"
mkdir -p "$AIM_REAL_HOME"

# Mock binaries
cat << 'MOCK' > "$MOCK_BIN/codex"
#!/bin/sh
if [ "$1" = "--version" ]; then
  echo "codex-cli 0.154.0"
  exit 0
fi
if [ "$1" = "mcp" ] && [ "$2" = "list" ]; then
  echo "NATIVE_MOCK_CODEX_MCP_LIST: $*"
  exit 0
fi
echo "mock codex invoked with: $*"
MOCK
chmod +x "$MOCK_BIN/codex"

cat << 'MOCK' > "$MOCK_BIN/claude"
#!/bin/sh
if [ "$1" = "--version" ]; then
  echo "claude 1.0.0"
  exit 0
fi
echo "mock claude invoked with: $*"
MOCK
chmod +x "$MOCK_BIN/claude"

cat << 'MOCK' > "$MOCK_BIN/agy"
#!/bin/sh
if [ "$1" = "--version" ]; then
  echo "agy 1.0.0"
  exit 0
fi
echo "mock agy invoked with: $*"
MOCK
chmod +x "$MOCK_BIN/agy"

# 3. Seed realistic sessions across multiple agents and profiles
echo "==> Seeding multi-agent sessions (Codex, Claude, Antigravity)..."

# 3a. Codex session with SQLite thread, rollout, and CWD
mkdir -p "$AIM_HOME/profiles/work/.codex/sessions/2026/09/22"
CODEX_ROLLOUT="$AIM_HOME/profiles/work/.codex/sessions/2026/09/22/rollout-01a0c7e1-e34c-7ca2-be3f-df5db5e94752.jsonl"
cat << 'JSONL' > "$CODEX_ROLLOUT"
{"event":"turn","turn_index":0,"content":"Implement user authentication flow"}
{"event":"turn","turn_index":1,"content":"Turn 1: Added JWT middleware"}
{"event":"turn","turn_index":2,"content":"Turn 2: Fixed token expiry handling"}
JSONL

CODEX_DB="$AIM_HOME/profiles/work/.codex/state_5.sqlite"
sqlite3 "$CODEX_DB" <<SQL
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
('01a0c7e1-e34c-7ca2-be3f-df5db5e94752', '$CODEX_ROLLOUT', 1727100000, 1727100500, 'cli', 'openai', '/Users/test/workspace/aim-backend', 'Auth Implementation Task', 'danger', 'manual');
SQL

# 3b. Claude session with project jsonl and slug CWD
CLAUDE_PROJ="$AIM_HOME/profiles/office/.claude/projects/-Users-test-workspace-aim-frontend"
mkdir -p "$CLAUDE_PROJ"
CLAUDE_SESS_ID="38769b32-e018-472c-98fc-1205166f2284"
cat << 'JSONL' > "$CLAUDE_PROJ/$CLAUDE_SESS_ID.jsonl"
{"type":"last-prompt","leafUuid":"leaf-0","sessionId":"38769b32-e018-472c-98fc-1205166f2284"}
{"type":"user","message":{"role":"user","content":"Build modern responsive navbar component"},"timestamp":"2026-09-23T12:00:00.000Z","sessionId":"38769b32-e018-472c-98fc-1205166f2284"}
{"type":"assistant","message":{"role":"assistant","content":"I will create the Navbar component with Tailwind CSS"},"timestamp":"2026-09-23T12:01:00.000Z","sessionId":"38769b32-e018-472c-98fc-1205166f2284"}
{"type":"user","message":{"role":"user","content":"Add mobile drawer menu to the navbar"},"timestamp":"2026-09-23T12:05:00.000Z","sessionId":"38769b32-e018-472c-98fc-1205166f2284"}
JSONL

# 3c. Antigravity session with brain logs, metadata, task.md, and conversation_summaries.db
AGY_SESS_ID="035e6b55-978f-4474-b25f-1c7059020323"
AGY_BRAIN="$AIM_HOME/profiles/work/.gemini/antigravity-cli/brain/$AGY_SESS_ID"
mkdir -p "$AGY_BRAIN/.system_generated/logs"
cat << 'JSONL' > "$AGY_BRAIN/.system_generated/logs/transcript.jsonl"
{"content":"Task checklist for resolving working directory display in sessions drawer"}
{"content":"Command Working Directory: /Users/test/workspace/my-project"}
JSONL
cat << 'JSON' > "$AGY_BRAIN/task.md.metadata.json"
{"goal":"Resolving working directory display and high-signal session preview card"}
JSON
cat << 'MD' > "$AGY_BRAIN/task.md"
# Task Checklist
- [x] Phase 1: Working Directory Display
- [x] Phase 2: Interface-level Architecture
- [ ] Phase 3: Live Verification
MD

AGY_DB="$AIM_HOME/profiles/work/.gemini/antigravity-cli/conversation_summaries.db"
sqlite3 "$AGY_DB" <<SQL
CREATE TABLE conversation_summaries (conversation_id TEXT PRIMARY KEY, title TEXT, preview TEXT, last_modified_time TEXT);
INSERT INTO conversation_summaries VALUES ('$AGY_SESS_ID', 'Task checklist for CWD', 'Resolving working directory display', '2026-09-23T12:00:00Z');
SQL

# 3d. Seed additional dummy sessions (12 more) to verify list windowing (total 15 sessions)
for i in $(seq 1 12); do
  P_ID="01a0dead-beef-0000-0000-$(printf '%012d' $i)"
  mkdir -p "$AIM_HOME/profiles/staging/.codex/sessions/2026/09/22"
  DUMMY_ROLLOUT="$AIM_HOME/profiles/staging/.codex/sessions/2026/09/22/rollout-$P_ID.jsonl"
  echo "{\"event\":\"turn\",\"turn_index\":0,\"content\":\"Dummy Task $i\"}" > "$DUMMY_ROLLOUT"
  STAGING_DB="$AIM_HOME/profiles/staging/.codex/state_5.sqlite"
  sqlite3 "$STAGING_DB" <<SQL
CREATE TABLE IF NOT EXISTS _sqlx_migrations (version BIGINT PRIMARY KEY, description TEXT NOT NULL, installed_on TIMESTAMP, success BOOLEAN, checksum BLOB, execution_time BIGINT);
INSERT OR IGNORE INTO _sqlx_migrations VALUES (1, 'init', CURRENT_TIMESTAMP, 1, X'00', 1);
CREATE TABLE IF NOT EXISTS threads (
	id TEXT PRIMARY KEY, rollout_path TEXT NOT NULL, created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL,
	source TEXT NOT NULL, model_provider TEXT NOT NULL, cwd TEXT NOT NULL, title TEXT NOT NULL,
	sandbox_policy TEXT NOT NULL, approval_mode TEXT NOT NULL
);
INSERT OR REPLACE INTO threads VALUES ('$P_ID', '$DUMMY_ROLLOUT', 1727000000, 1727000000 + $i, 'cli', 'openai', '/Users/test/workspace/staging-service', 'Dummy Task $i', 'danger', 'manual');
SQL
done

# Seed MCP configs for interception tests
cat << 'CODEXCFG' > "$AIM_HOME/profiles/work/.codex/config.toml"
[mcp_servers.playwright]
command = "npx"
args = ["@playwright/mcp@latest"]

[mcp_servers.atlassian-oauth]
url = "https://mcp.atlassian.com/v2/mcp"
CODEXCFG

echo "✔ Sandbox seeded with 15 sessions across 3 profiles"

# ==============================================================================
# PHASE 1: Live Interactive Sessions Explorer (PTY + Bubble Tea Rendering)
# ==============================================================================
echo ""
echo "=== Phase 1: Live Interactive Sessions Explorer (PTY) ==="

python3 - << 'PYTEST'
import pty, os, subprocess, time, struct, fcntl, termios, re, sys

aim_bin = os.environ["AIM_BIN"]
master, slave = pty.openpty()
# 110 columns x 35 rows standard terminal
fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 35, 110, 0, 0))

start_time = time.time()
proc = subprocess.Popen([aim_bin, "sessions"], stdin=slave, stdout=slave, stderr=slave, close_fds=True)
os.close(slave)

# Read output until rendered, measure latency
out = b""
rendered = False
while time.time() - start_time < 3.0:
    try:
        data = os.read(master, 2048)
        if data:
            out += data
            if b"Sessions Explorer" in out and b"PROFILE" in out:
                rendered = True
                break
    except OSError:
        break

elapsed = time.time() - start_time
print(f"  [Latency Check] Interactive Sessions Explorer rendered in {elapsed:.3f}s")
if not rendered:
    print("FAIL: Sessions Explorer did not render within 3.0s")
    sys.exit(1)
if elapsed > 2.0:
    print(f"FAIL: Latency regression detected! Render took {elapsed:.3f}s > 2.0s")
    sys.exit(1)

# Send arrow down to test cursor movement, then quit
os.write(master, b"\x1b[B")
time.sleep(0.1)
os.write(master, b"q")

# Read remainder
while True:
    try:
        data = os.read(master, 2048)
        if not data:
            break
        out += data
    except OSError:
        break

proc.wait()
os.close(master)

if proc.returncode != 0:
    print(f"FAIL: aim sessions exited with code {proc.returncode}, expected 0")
    sys.exit(1)

clean = re.sub(r"\x1b\[[0-9;]*[a-zA-Z]", "", out.decode("utf-8", errors="replace"))

# Assertions
assert "Sessions Explorer" in clean, "Missing 'Sessions Explorer' header"
assert "[0] All" in clean or "[0-3/Tab]" in clean, "Missing agent filter tabs"
assert "DIR" in clean, "Table header missing 'DIR' column"
assert "staging-service" in clean or "aim-backend" in clean or "aim-frontend" in clean or "my-project" in clean, "Table row missing workspace directory"
assert "Workspace:" in clean, "Preview card missing 'Workspace:' section"
assert "Goal:" in clean, "Preview card missing 'Goal:' section"

# Deduplication check: verify no old duplicate hint in preview card
assert "enter resume" not in clean or "[enter] Resume" in clean, "Footer bar should use unified [enter] Resume format"
assert "↑/↓ select • enter resume" not in clean, "FAIL: Found obsolete duplicate hint '↑/↓ select • enter resume' inside preview card!"

# Windowing check: verify table indicates windowing and does not blast 15 rows
assert "showing 1-" in clean or "(showing" in clean, "Missing windowing indicator (e.g. showing 1-8)"

print("  ✔ Instant latency verified (< 2s)")
print("  ✔ Header, command shortcuts, and table DIR column verified")
print("  ✔ Preview card contains Workspace & Goal")
print("  ✔ Deduplication verified: no redundant hints inside preview box")
print("  ✔ Windowing verified: list is capped and windowed")
print("  ✔ Clean exit on 'q'")
PYTEST

# ==============================================================================
# PHASE 2: Live Main TUI Dashboard & Drawer Navigation (PTY)
# ==============================================================================
echo ""
echo "=== Phase 2: Live Main TUI Dashboard & Drawer Navigation (PTY) ==="

python3 - << 'PYTEST'
import pty, os, subprocess, time, struct, fcntl, termios, re, sys

aim_bin = os.environ["AIM_BIN"]
master, slave = pty.openpty()
fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 35, 110, 0, 0))

proc = subprocess.Popen([aim_bin], stdin=slave, stdout=slave, stderr=slave, close_fds=True)
os.close(slave)

time.sleep(0.4)
# Press 's' to open Sessions Drawer
os.write(master, b"s")
time.sleep(0.4)

# Read output
out = b""
for _ in range(5):
    try:
        data = os.read(master, 2048)
        if data:
            out += data
    except OSError:
        break

# Press 'q' to close drawer, then 'q' to exit TUI
os.write(master, b"q")
time.sleep(0.1)
os.write(master, b"q")

while True:
    try:
        data = os.read(master, 2048)
        if not data:
            break
        out += data
    except OSError:
        break

proc.wait()
os.close(master)

clean = re.sub(r"\x1b\[[0-9;]*[a-zA-Z]", "", out.decode("utf-8", errors="replace"))
assert "Sessions Explorer" in clean or "SESSION ID" in clean, "Pressing 's' did not open Sessions Drawer"

print("  ✔ Main TUI opened cleanly")
print("  ✔ 's' key successfully toggled Sessions Drawer")
print("  ✔ Exited cleanly")
PYTEST

# ==============================================================================
# PHASE 3: CLI Non-Interactive Plain Table Formatting
# ==============================================================================
echo ""
echo "=== Phase 3: CLI Non-Interactive Plain Table Formatting ==="

PLAIN_OUT=$("$AIM_BIN" sessions --plain)
echo "$PLAIN_OUT" | grep -q "RECENT SESSIONS" || { echo "FAIL: Missing RECENT SESSIONS header in --plain"; exit 1; }
echo "$PLAIN_OUT" | grep -q "DIR" || { echo "FAIL: Missing DIR column in --plain table"; exit 1; }
echo "$PLAIN_OUT" | grep -q "01a0c7e1" || { echo "FAIL: Missing codex session 01a0c7e1"; exit 1; }
echo "$PLAIN_OUT" | grep -q "aim-backend" || { echo "FAIL: Missing workspace path in DIR column"; exit 1; }
echo "  ✔ 'aim sessions --plain' formatted tables with DIR column"

# Agent filtering
CODEX_FILTER=$("$AIM_BIN" sessions codex --plain)
echo "$CODEX_FILTER" | grep -q "codex" || { echo "FAIL: Missing codex in filtered output"; exit 1; }
if echo "$CODEX_FILTER" | grep -q "claude"; then
  echo "FAIL: Claude session appeared in 'aim sessions codex'"
  exit 1
fi
echo "  ✔ Agent filtering ('aim sessions codex --plain') works accurately"

# Profile filtering
WORK_FILTER=$("$AIM_BIN" sessions -p work --plain)
echo "$WORK_FILTER" | grep -q "work" || { echo "FAIL: Missing work profile in filtered output"; exit 1; }
if echo "$WORK_FILTER" | grep -q "office"; then
  echo "FAIL: Office profile appeared in 'aim sessions -p work'"
  exit 1
fi
echo "  ✔ Profile filtering ('aim sessions -p work --plain') works accurately"

# ==============================================================================
# PHASE 4: CLI Structured JSON Output
# ==============================================================================
echo ""
echo "=== Phase 4: CLI Structured JSON Output ==="

JSON_OUT=$("$AIM_BIN" sessions --json)
python3 -c '
import json, sys
data = json.loads(sys.argv[1])
assert isinstance(data, list), "Expected JSON array"
assert len(data) >= 3, f"Expected at least 3 sessions, got {len(data)}"
found_cwd = False
found_goal = False
for s in data:
    assert "id" in s and "short_id" in s and "agent" in s and "profile" in s
    if s.get("cwd"):
        found_cwd = True
    if s.get("goal"):
        found_goal = True
assert found_cwd, "No session contained a populated cwd field in JSON"
assert found_goal, "No session contained a populated goal field in JSON"
print(f"  ✔ JSON output valid: {len(data)} sessions parsed with populated cwd and goal fields")
' "$JSON_OUT"

# ==============================================================================
# PHASE 5: Transparent MCP List Interception & Passthrough
# ==============================================================================
echo ""
echo "=== Phase 5: Transparent MCP List Interception & Passthrough ==="

MCP_TABLE=$("$AIM_BIN" run codex work mcp list)
echo "$MCP_TABLE" | grep -q "=== Configured MCP Servers (codex: work) ===" || {
  echo "FAIL: Expected styled table banner for 'aim run codex work mcp list'"
  echo "$MCP_TABLE"
  exit 1
}
echo "$MCP_TABLE" | grep -q "playwright" || { echo "FAIL: Missing playwright in MCP table"; exit 1; }
echo "$MCP_TABLE" | grep -q "atlassian-oauth" || { echo "FAIL: Missing atlassian-oauth in MCP table"; exit 1; }
if echo "$MCP_TABLE" | grep -q "NATIVE_MOCK_CODEX_MCP_LIST"; then
  echo "FAIL: Plain 'mcp list' should have been intercepted, but native mock ran"
  exit 1
fi
echo "  ✔ 'aim run codex work mcp list' was transparently intercepted into styled table"

# Passthrough on --json
MCP_JSON=$("$AIM_BIN" run codex work mcp list --json)
echo "$MCP_JSON" | grep -q "NATIVE_MOCK_CODEX_MCP_LIST" || {
  echo "FAIL: 'mcp list --json' was NOT passed through to native agent"
  echo "$MCP_JSON"
  exit 1
}
echo "  ✔ 'aim run codex work mcp list --json' cleanly passed through to agent CLI"

# ==============================================================================
# PHASE 6: Redundant Command Removal Guardrails
# ==============================================================================
echo ""
echo "=== Phase 6: Redundant Command Removal Guardrails ==="

# 1. aim mcp
MCP_ERR=$("$AIM_BIN" mcp 2>&1 || true)
if echo "$MCP_ERR" | grep -q "unknown command \"mcp\""; then
  echo "  ✔ 'aim mcp' correctly rejected as unknown command"
else
  echo "FAIL: 'aim mcp' was not rejected as unknown command: $MCP_ERR"
  exit 1
fi

# 2. aim clone
CLONE_ERR=$("$AIM_BIN" clone 2>&1 || true)
if echo "$CLONE_ERR" | grep -q "unknown command \"clone\""; then
  echo "  ✔ 'aim clone' correctly rejected as unknown command"
else
  echo "FAIL: 'aim clone' was not rejected as unknown command: $CLONE_ERR"
  exit 1
fi

# 3. aim sessions show
# When show is passed, aim sessions treats it as [agent] filter which will find no agent or no sessions, NOT a subcommand
SHOW_OUT=$("$AIM_BIN" sessions show --plain 2>&1 || true)
if echo "$SHOW_OUT" | grep -qi "no sessions found"; then
  echo "  ✔ 'aim sessions show' is not a registered subcommand"
else
  echo "FAIL: 'aim sessions show' behaved unexpectedly: $SHOW_OUT"
  exit 1
fi

# 4. aim sessions import
IMPORT_OUT=$("$AIM_BIN" sessions import --plain 2>&1 || true)
if echo "$IMPORT_OUT" | grep -qi "no sessions found"; then
  echo "  ✔ 'aim sessions import' is not a registered subcommand"
else
  echo "FAIL: 'aim sessions import' behaved unexpectedly: $IMPORT_OUT"
  exit 1
fi

echo ""
echo "========================================================================"
echo "  ALL E2E LIVE TESTS PASSED! (0 manual checks required)                  "
echo "========================================================================"
