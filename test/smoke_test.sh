#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AIM_BIN="$REPO_ROOT/aim"

(cd "$REPO_ROOT" && go build -o aim ./cmd/aim)

TEST_AIM_HOME=$(mktemp -d)
MOCK_BIN=$(mktemp -d)
trap "rm -rf '$TEST_AIM_HOME' '$MOCK_BIN'" EXIT

export AIM_HOME="$TEST_AIM_HOME"
export AIM_REAL_HOME="$TEST_AIM_HOME/fake_home"
export AIM_AUTO_CREATE=1
mkdir -p "$TEST_AIM_HOME/fake_home/.gemini/antigravity-cli/conversations"
mkdir -p "$TEST_AIM_HOME/fake_home/.gemini/antigravity-cli/brain"

cat << 'MOCK' > "$MOCK_BIN/agy"
#!/bin/sh
echo "mock agy executed with args: $@"
if [ -n "${MOCK_CUSTOM_ENV:-}" ]; then
  echo "mock agy received env: $MOCK_CUSTOM_ENV"
fi
MOCK
chmod +x "$MOCK_BIN/agy"

cat << 'MOCK' > "$MOCK_BIN/gemini"
#!/bin/sh
echo "mock gemini executed with args: $@"
MOCK
chmod +x "$MOCK_BIN/gemini"

cat << 'MOCK' > "$MOCK_BIN/codex"
#!/bin/sh
if [ "$1" = "--version" ]; then
  echo "codex-cli 0.154.0"
  exit 0
fi
echo "mock codex executed with args: $@"
MOCK
chmod +x "$MOCK_BIN/codex"

cat << 'MOCK' > "$MOCK_BIN/claude"
#!/bin/sh
if [ "$1" = "--version" ]; then
  echo "claude 1.0.0"
  exit 0
fi
if [ "$1" = "auth" ] && [ "$2" = "login" ]; then
  echo "mock claude login successful"
  exit 0
fi
echo "mock claude executed with args: $@"
MOCK
chmod +x "$MOCK_BIN/claude"

export PATH="$MOCK_BIN:$PATH"

echo "=== 1. Testing help & version ==="
"$AIM_BIN" --version
"$AIM_BIN" --help > /dev/null

echo "=== 2. Testing initial empty listing ==="
"$AIM_BIN" list agy | grep "No profiles found"
"$AIM_BIN" list gemini | grep "No profiles found"
"$AIM_BIN" list codex | grep "No profiles found"
"$AIM_BIN" list claude | grep "No profiles found"

echo "=== 3. Testing dotfile isolation & running agy ==="
mkdir -p "$TEST_AIM_HOME/fake_home/.agents/skills"
echo "[user] name = SmokeTest" > "$TEST_AIM_HOME/fake_home/.gitconfig"
mkdir -p "$TEST_AIM_HOME/fake_home/.gemini/config/plugins/test-plugin"
echo '{"name":"test-plugin"}' > "$TEST_AIM_HOME/fake_home/.gemini/config/plugins/test-plugin/plugin.json"
echo '{"imports":[{"name":"test-plugin"}]}' > "$TEST_AIM_HOME/fake_home/.gemini/config/import_manifest.json"
HOME="$TEST_AIM_HOME/fake_home" "$AIM_BIN" run agy smoke_profile -- echo "isolated"

[ -L "$TEST_AIM_HOME/profiles/smoke_profile/.gitconfig" ]
[ -L "$TEST_AIM_HOME/profiles/smoke_profile/.agents" ]
[ -L "$TEST_AIM_HOME/profiles/smoke_profile/.gemini/config/plugins" ]
[ -L "$TEST_AIM_HOME/profiles/smoke_profile/.gemini/config/import_manifest.json" ]
echo "Dotfile symlink and plugin bridging OK!"

echo "=== 4. Testing agent association after run ==="
# aim list agy should show smoke_profile
"$AIM_BIN" list agy | grep "smoke_profile"

# Verify agent-first shorthand is rejected
if "$AIM_BIN" agy list 2>/dev/null; then
  echo "Error: agent-first shorthand 'aim agy list' should fail"
  exit 1
fi

# aim list gemini should NOT show smoke_profile
LIST_GEMINI="$("$AIM_BIN" list gemini)"
if echo "$LIST_GEMINI" | grep -q "smoke_profile"; then
  echo "Error: gemini list should not show smoke_profile before gemini is run"
  exit 1
fi
"$AIM_BIN" list gemini | grep "No profiles found"

echo "=== 5. Testing multi-agent attachment ==="
# Run gemini with smoke_profile
"$AIM_BIN" run gemini smoke_profile -- echo "gemini running"

# Run codex with smoke_profile
"$AIM_BIN" run codex smoke_profile -- echo "codex running"
[ -d "$TEST_AIM_HOME/profiles/smoke_profile/.codex" ]

# agy, gemini, and codex list should now show smoke_profile
"$AIM_BIN" list agy | grep "smoke_profile"
"$AIM_BIN" list gemini | grep "smoke_profile"
"$AIM_BIN" list codex | grep "smoke_profile"

# Verify agent-first shorthand is rejected
if "$AIM_BIN" gemini list 2>/dev/null; then
  echo "Error: agent-first shorthand 'aim gemini list' should fail"
  exit 1
fi
if "$AIM_BIN" codex list 2>/dev/null; then
  echo "Error: agent-first shorthand 'aim codex list' should fail"
  exit 1
fi

# aim list (all) displays smoke_profile [agy, gemini, codex]
"$AIM_BIN" list | grep -E "smoke_profile \[agy, gemini, codex\]"
echo "Multi-agent tag display OK!"

echo "=== 6. Testing shell completion ==="
# 1. aim completion zsh emits #compdef aim and _describe
"$AIM_BIN" completion zsh | grep "#compdef aim"
"$AIM_BIN" completion zsh | grep "_describe"

# 2. aim completion bash emits complete -o default -F __start_aim aim
"$AIM_BIN" completion bash | grep "__start_aim"

# 3. aim completion fish emits complete -c aim
"$AIM_BIN" completion fish | grep "complete -c aim"

# 4. aim __complete emits root subcommands including completion:
"$AIM_BIN" __complete | grep "^completion"
"$AIM_BIN" __complete | grep "^run"

# 5. aim __complete run emits agy, gemini, and codex
"$AIM_BIN" __complete run | grep "agy"
"$AIM_BIN" __complete run | grep "gemini"
"$AIM_BIN" __complete run | grep "codex"
"$AIM_BIN" __complete run | grep "claude"
"$AIM_BIN" __complete clone | grep "agy"
"$AIM_BIN" __complete clone | grep "gemini"
"$AIM_BIN" __complete clone | grep "codex"
"$AIM_BIN" __complete clone | grep "claude"
"$AIM_BIN" __complete mv | grep "agy"
"$AIM_BIN" __complete mv | grep "gemini"
"$AIM_BIN" __complete mv | grep "codex"
"$AIM_BIN" __complete mv | grep "claude"

# 6. aim __complete run emits configured profiles (e.g. smoke_profile)
"$AIM_BIN" __complete run agy | grep "smoke_profile"
"$AIM_BIN" __complete run gemini | grep "smoke_profile"
"$AIM_BIN" __complete run codex | grep "smoke_profile"
"$AIM_BIN" __complete clone agy | grep "smoke_profile"
"$AIM_BIN" __complete clone codex | grep "smoke_profile"
"$AIM_BIN" __complete mv codex | grep "smoke_profile"

# Shorthand completion rejected
SHORTHAND_COMP="$("$AIM_BIN" __complete agy run 2>/dev/null || true)"
if echo "$SHORTHAND_COMP" | grep -q "smoke_profile"; then
  echo "Error: shorthand completion 'aim __complete agy run' should not emit profiles"
  exit 1
fi
echo "Shell completion OK!"

echo "=== 7. Testing doctor ==="
# aim doctor without args iterates over registered agents
DOCTOR_OUT="$("$AIM_BIN" doctor)"
echo "$DOCTOR_OUT" | grep "AIM Doctor Diagnostics"
echo "$DOCTOR_OUT" | grep "Profile: smoke_profile"
echo "$DOCTOR_OUT" | grep "Binary: Found agy"
echo "$DOCTOR_OUT" | grep "Gemini adapter registered"
echo "$DOCTOR_OUT" | grep "Codex installed"

# aim doctor codex specifically checks codex
"$AIM_BIN" doctor codex | grep "Codex installed"
! "$AIM_BIN" doctor codex | grep -i "Gemini"

# aim doctor agy specifically checks agy
"$AIM_BIN" doctor agy | grep "Binary: Found agy"
! "$AIM_BIN" doctor agy | grep -i "Gemini"

echo "=== 8. Testing usage command ==="
"$AIM_BIN" usage
"$AIM_BIN" usage agy

# Assert aim list displays usage badge for smoke_profile after usage query
"$AIM_BIN" list | grep "smoke_profile" | grep "\[no credentials\]"
"$AIM_BIN" list agy | grep "smoke_profile" | grep "\[no credentials\]"

echo "=== 9. Testing usage JSON output ==="
JSON_OUT=$("$AIM_BIN" usage agy smoke_profile --json)
echo "$JSON_OUT" | grep -q '"agent": "agy"' || { echo "FAIL: missing agent in usage json"; exit 1; }
echo "$JSON_OUT" | grep -q '"profile": "smoke_profile"' || { echo "FAIL: missing profile in usage json"; exit 1; }

echo "=== 10. Testing completion for usage ==="
USAGE_COMP="$("$AIM_BIN" __complete usage)"
echo "$USAGE_COMP" | grep -q "agy" || { echo "FAIL: missing agy in usage completion: $USAGE_COMP"; exit 1; }

echo "=== 11. Testing conversation bridging & continuation flags ==="

# Prepare conversation in shared directory
touch "$TEST_AIM_HOME/fake_home/.gemini/antigravity-cli/conversations/smoke-conv-001.db"

# Run default continuation
RUN_CONT="$("$AIM_BIN" run agy smoke_profile --continue)"
echo "$RUN_CONT" | grep "mock agy executed with args: --continue"

# Run specific session
RUN_SESS="$("$AIM_BIN" run agy smoke_profile --conversation smoke-conv-001)"
echo "$RUN_SESS" | grep "mock agy executed with args: --conversation smoke-conv-001"

# Verify symlinks were created in profile directory
[ -L "$TEST_AIM_HOME/profiles/smoke_profile/.gemini/antigravity-cli/conversations" ]
[ -L "$TEST_AIM_HOME/profiles/smoke_profile/.gemini/antigravity-cli/brain" ]
[ -f "$TEST_AIM_HOME/profiles/smoke_profile/.gemini/antigravity-cli/conversations/smoke-conv-001.db" ]

echo "Conversation bridging and continuation flags OK!"

echo "=== 12. Testing profile cloning and config overrides ==="
# Clone smoke_profile to clone_prof for agy
"$AIM_BIN" clone agy smoke_profile clone_prof
[ -d "$TEST_AIM_HOME/profiles/clone_prof" ]
[ -L "$TEST_AIM_HOME/profiles/clone_prof/.gitconfig" ]
"$AIM_BIN" list agy | grep "clone_prof"

# Test sensitive token exclusion during clone
mkdir -p "$TEST_AIM_HOME/profiles/smoke_profile/.gemini/antigravity-cli"
echo "secret" > "$TEST_AIM_HOME/profiles/smoke_profile/.gemini/antigravity-cli/antigravity-oauth-token"
"$AIM_BIN" clone agy smoke_profile clone_token_check
[ ! -f "$TEST_AIM_HOME/profiles/clone_token_check/.gemini/antigravity-cli/antigravity-oauth-token" ]
"$AIM_BIN" remove agy clone_token_check
rm -f "$TEST_AIM_HOME/profiles/smoke_profile/.gemini/antigravity-cli/antigravity-oauth-token"

# Configure custom env & args for clone_prof in config.json
python3 -c "
import json
p = '$TEST_AIM_HOME/config.json'
with open(p) as f:
    cfg = json.load(f)
cfg['profiles']['clone_prof']['env'] = {'MOCK_CUSTOM_ENV': 'smoke_override_active'}
cfg['profiles']['clone_prof']['args'] = ['--profile-flag']
with open(p, 'w') as f:
    json.dump(cfg, f, indent=2)
"

# Run clone_prof and verify env and args injection
RUN_OUT="$("$AIM_BIN" run agy clone_prof -- extra_cli_arg)"
echo "$RUN_OUT" | grep "mock agy executed with args: --profile-flag extra_cli_arg"
echo "$RUN_OUT" | grep "mock agy received env: smoke_override_active"

# Doctor verifies config overrides reported
DOCTOR_CLONE_OUT="$("$AIM_BIN" doctor agy)"
echo "$DOCTOR_CLONE_OUT" | grep "1 custom env var(s) configured"
echo "$DOCTOR_CLONE_OUT" | grep "1 custom launch arg(s) configured"

# Remove clone_prof
"$AIM_BIN" remove agy clone_prof
[ ! -d "$TEST_AIM_HOME/profiles/clone_prof" ]
echo "Profile clone and config overrides OK!"

echo "=== 12b. Verifying rename is not a CLI command (TUI action only) ==="
if "$AIM_BIN" rename smoke_profile smoke_renamed 2>/dev/null; then
  echo "Error: 'aim rename' should not be a CLI command (TUI action only)"
  exit 1
fi

echo "=== 12c. Testing agent account move (aim mv) ==="
# Verify 'aim move' alias does not exist (aim mv only)
if "$AIM_BIN" move codex smoke_profile mv_target 2>/dev/null; then
  echo "Error: 'aim move' should not exist ('aim mv' only)"
  exit 1
fi

echo '{"token":"smoke_codex_token"}' > "$TEST_AIM_HOME/profiles/smoke_profile/.codex/auth.json"

# Move codex from smoke_profile to mv_target
"$AIM_BIN" mv codex smoke_profile mv_target

# Assert mv_target has codex profile, credentials directory, and the file moved
[ -d "$TEST_AIM_HOME/profiles/mv_target/.codex" ]
[ -f "$TEST_AIM_HOME/profiles/mv_target/.codex/auth.json" ]
grep -q "smoke_codex_token" "$TEST_AIM_HOME/profiles/mv_target/.codex/auth.json"
"$AIM_BIN" list codex | grep "mv_target"

# smoke_profile should no longer have codex
if "$AIM_BIN" list codex | grep -q "smoke_profile"; then
  echo "Error: smoke_profile should not have codex after moving to mv_target"
  exit 1
fi
# But smoke_profile still has agy and gemini!
"$AIM_BIN" list agy | grep "smoke_profile"
"$AIM_BIN" list gemini | grep "smoke_profile"

# Move codex back to smoke_profile
"$AIM_BIN" mv codex mv_target smoke_profile

# Clean up dummy auth.json so subsequent steps retain expected empty credentials state
rm -f "$TEST_AIM_HOME/profiles/smoke_profile/.codex/auth.json"

# Assert mv_target was cleaned up (as it had only codex)
[ ! -d "$TEST_AIM_HOME/profiles/mv_target" ]
"$AIM_BIN" list codex | grep "smoke_profile"
echo "aim mv agent account OK!"

echo "=== 13. Testing partial removal (gemini and codex) ==="
"$AIM_BIN" remove gemini smoke_profile
# Assert directory still exists
[ -d "$TEST_AIM_HOME/profiles/smoke_profile" ]
echo "Profile directory preserved after partial removal OK!"

# Assert agy and codex still show smoke_profile
"$AIM_BIN" list agy | grep "smoke_profile"
"$AIM_BIN" list codex | grep "smoke_profile"

# Assert gemini no longer displays smoke_profile
LIST_GEMINI_AFTER="$("$AIM_BIN" list gemini)"
if echo "$LIST_GEMINI_AFTER" | grep -q "smoke_profile"; then
  echo "Error: gemini list should not show smoke_profile after gemini removal"
  exit 1
fi
"$AIM_BIN" list gemini | grep "No profiles found"

# Assert aim list shows smoke_profile [agy, codex]
"$AIM_BIN" list | grep -E "smoke_profile \[agy, codex\]"

# Now remove codex as well
"$AIM_BIN" remove codex smoke_profile
[ -d "$TEST_AIM_HOME/profiles/smoke_profile" ]
"$AIM_BIN" list codex | grep "No profiles found"

# Assert aim list shows smoke_profile [agy]
"$AIM_BIN" list | grep -E "smoke_profile \[agy\]"
echo "Partial removal state OK!"

echo "=== 14. Testing full removal (agy) ==="
"$AIM_BIN" remove agy smoke_profile
# Assert directory is deleted
[ ! -d "$TEST_AIM_HOME/profiles/smoke_profile" ]
echo "Profile directory deleted after full removal OK!"

# Assert listing is empty
"$AIM_BIN" list | grep "No profiles found"

# Assert bare doctor when no profiles exist iterates over registered agents
DOCTOR_EMPTY="$("$AIM_BIN" doctor)"
echo "$DOCTOR_EMPTY" | grep "No profiles configured for agent \"agy\""
echo "$DOCTOR_EMPTY" | grep "No profiles configured for agent \"gemini\""
echo "$DOCTOR_EMPTY" | grep "No profiles configured for agent \"codex\""
echo "$DOCTOR_EMPTY" | grep "No profiles configured for agent \"claude\""

echo "=== 15. Testing sessions, resume & import CLI commands ==="
# Test sessions help
"$AIM_BIN" sessions --help > /dev/null
"$AIM_BIN" chats --help > /dev/null

# Test sessions listing empty & json output
"$AIM_BIN" sessions | grep -i "no sessions found"
"$AIM_BIN" sessions --json | grep "\[\]"

# Test resume help & validation
"$AIM_BIN" resume --help > /dev/null
if "$AIM_BIN" resume 2>/dev/null; then
  echo "Error: bare 'aim resume' should fail"
  exit 1
fi

# Test sessions import help & validation
"$AIM_BIN" sessions import --help > /dev/null
if "$AIM_BIN" sessions import 2>/dev/null; then
  echo "Error: bare 'aim sessions import' should fail"
  exit 1
fi
echo "Sessions, resume & import CLI commands OK!"

echo "=== 16. Testing Claude Code adapter (run, doctor, sessions, resume) ==="
# 1. Test aim run claude testprof
RUN_CLAUDE="$("$AIM_BIN" run claude testprof -- echo "hello claude")"
echo "$RUN_CLAUDE" | grep "mock claude executed with args: echo hello claude"
[ -d "$TEST_AIM_HOME/profiles/testprof/.claude" ]
"$AIM_BIN" list claude | grep "testprof"

# 2. Test aim doctor claude testprof
DOCTOR_CLAUDE="$("$AIM_BIN" doctor claude)"
echo "$DOCTOR_CLAUDE" | grep "Claude installed (claude 1.0.0"
echo "$DOCTOR_CLAUDE" | grep "Storage directory"

# 3. Seed a session for claude in testprof
PROJ_SLUG="-Users-test-projects-aim"
mkdir -p "$TEST_AIM_HOME/profiles/testprof/.claude/projects/$PROJ_SLUG"
CLAUDE_SESS_ID="c24699d0-820f-435f-b4e4-eaf8440311b4"
cat << 'JSONL' > "$TEST_AIM_HOME/profiles/testprof/.claude/projects/$PROJ_SLUG/$CLAUDE_SESS_ID.jsonl"
{"type":"last-prompt","leafUuid":"leaf-1","sessionId":"c24699d0-820f-435f-b4e4-eaf8440311b4"}
{"type":"user","message":{"role":"user","content":"Claude smoke test session"},"timestamp":"2026-09-23T12:00:00.000Z","sessionId":"c24699d0-820f-435f-b4e4-eaf8440311b4"}
JSONL

# 4. Test aim sessions claude
SESSIONS_CLAUDE="$("$AIM_BIN" sessions claude -p testprof)"
echo "$SESSIONS_CLAUDE" | grep "c24699d0"
echo "$SESSIONS_CLAUDE" | grep "Claude smoke test session"

# 5. Test aim resume claude testprof c24699d0
RESUME_CLAUDE="$("$AIM_BIN" resume claude testprof c24699d0)"
echo "$RESUME_CLAUDE" | grep "mock claude executed with args: --resume c24699d0-820f-435f-b4e4-eaf8440311b4"

# 6. Clean up testprof
"$AIM_BIN" remove claude testprof
[ ! -d "$TEST_AIM_HOME/profiles/testprof" ]
echo "Claude Code adapter smoke tests OK!"

echo "ALL SMOKE TESTS PASSED!"
