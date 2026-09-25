#!/usr/bin/env bash
set -euo pipefail

# Real AIM -> /bin/sh -> AIM subprocess checks. No agent binaries are mocked or
# invoked here; only macOS Keychain access is stubbed to protect host credentials.
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_ROOT=$(mktemp -d)
trap 'rm -rf -- "$TEST_ROOT"' EXIT
export AIM_TEST_BINARY="$TEST_ROOT/aim"
(cd "$REPO_ROOT" && go build -o "$AIM_TEST_BINARY" ./cmd/aim)

export AIM_AUTO_CREATE=1
export SHELL=/bin/sh
export AIM_TEST_INHERITED=benign-value
mkdir -p "$TEST_ROOT/bin"
cat > "$TEST_ROOT/bin/security" <<'MOCK'
#!/bin/sh
exit 44
MOCK
chmod +x "$TEST_ROOT/bin/security"
export PATH="$TEST_ROOT/bin:$PATH"

for mode in custom legacy xdg default; do
unset AIM_HOME AIM_CONFIG_DIR AIM_DATA_DIR AIM_CACHE_DIR AIM_STATE_DIR
export AIM_REAL_HOME="$TEST_ROOT/$mode/host"
export XDG_CONFIG_HOME="$TEST_ROOT/$mode/config"
export XDG_DATA_HOME="$TEST_ROOT/$mode/data"
export XDG_CACHE_HOME="$TEST_ROOT/$mode/cache"
export XDG_STATE_HOME="$TEST_ROOT/$mode/state"
case "$mode" in
  custom) export AIM_HOME="$TEST_ROOT/custom store" ;;
  legacy) mkdir -p "$AIM_REAL_HOME/.aim" ;;
  default) unset XDG_CONFIG_HOME XDG_DATA_HOME XDG_CACHE_HOME XDG_STATE_HOME ;;
esac
case "$mode" in
  custom) store="$AIM_HOME"; config_dir="$store" ;;
  legacy) store="$AIM_REAL_HOME/.aim"; config_dir="$store" ;;
  xdg) store="$XDG_DATA_HOME/aim"; config_dir="$XDG_CONFIG_HOME/aim" ;;
  default)
    case "$(uname -s)" in
      Darwin) store="$AIM_REAL_HOME/Library/Application Support/aim"; config_dir="$store" ;;
      *) store="$AIM_REAL_HOME/.local/share/aim"; config_dir="$AIM_REAL_HOME/.config/aim" ;;
    esac ;;
esac
export AIM_TEST_EXPECTED_STORE="$store"
export AIM_TEST_EXPECTED_HOME=""
case "$mode" in custom|legacy) export AIM_TEST_EXPECTED_HOME="$store" ;; esac
mkdir -p "$AIM_REAL_HOME" "$config_dir"
cat > "$config_dir/config.json" <<'CONFIG'
{
  "profiles": {
    "plain": {"agents": ["agy", "gemini", "codex", "claude"]},
    "overridden": {
      "agents": ["agy", "gemini", "codex", "claude"],
      "env": {"AIM_TEST_INHERITED": "profile-value", "AIM_SESSION_ID": "explicit-session"},
      "args": ["--agent-only-argument"]
    }
  }
}
CONFIG

for agent in agy gemini codex claude; do
  for profile in plain overridden; do
    echo "Live shell: $mode/$agent/$profile"
    export AIM_TEST_EXPECTED_AGENT="$agent"
    export AIM_TEST_EXPECTED_PROFILE="$profile"
    # Inject stale managed values and fake tokens only into this test process.
    # Never print ambient real credentials in diagnostics.
    env HOME="$AIM_REAL_HOME" AIM_AGENT=stale AIM_PROFILE=stale AIM_SESSION_ID=stale \
      GEMINI_CLI_HOME=stale CODEX_HOME=stale CLAUDE_CONFIG_DIR=stale \
      SSH_CONNECTION=stale SSH_CLIENT=stale SSH_TTY=stale \
      CLAUDE_CODE_OAUTH_TOKEN=fake ANTHROPIC_API_KEY=fake CLAUDE_CODE_OAUTH_REFRESH_TOKEN=fake \
      "$AIM_TEST_BINARY" shell "$agent" "$profile" <<'CHECK'
set -eu
[ "${AIM_HOME-}" = "$AIM_TEST_EXPECTED_HOME" ] || { echo 'FAIL: storage mode changed'; exit 1; }
[ "$HOME" = "$AIM_TEST_EXPECTED_STORE/profiles/$AIM_TEST_EXPECTED_PROFILE" ]
[ "$AIM_AGENT" = "$AIM_TEST_EXPECTED_AGENT" ]
[ "$AIM_PROFILE" = "$AIM_TEST_EXPECTED_PROFILE" ]
[ -z "${SSH_CONNECTION-}${SSH_CLIENT-}${SSH_TTY-}" ]
[ -z "${CLAUDE_CODE_OAUTH_TOKEN-}${ANTHROPIC_API_KEY-}${CLAUDE_CODE_OAUTH_REFRESH_TOKEN-}" ]
case "$AIM_AGENT" in
  agy) [ -z "${GEMINI_CLI_HOME-}${CODEX_HOME-}${CLAUDE_CONFIG_DIR-}" ] ;;
  gemini) [ "$GEMINI_CLI_HOME" = "$HOME/.gemini" ]; [ -z "${CODEX_HOME-}${CLAUDE_CONFIG_DIR-}" ] ;;
  codex) [ "$CODEX_HOME" = "$HOME/.codex" ]; [ -z "${GEMINI_CLI_HOME-}${CLAUDE_CONFIG_DIR-}" ] ;;
  claude) [ "$CLAUDE_CONFIG_DIR" = "$HOME/.claude" ]; [ -z "${GEMINI_CLI_HOME-}${CODEX_HOME-}" ] ;;
esac
if [ "$AIM_PROFILE" = overridden ]; then
  [ "$AIM_TEST_INHERITED" = profile-value ]
  [ "$AIM_SESSION_ID" = explicit-session ]
else
  [ "$AIM_TEST_INHERITED" = benign-value ]
  [ -z "${AIM_SESSION_ID-}" ]
fi
# The nested executable must find the same on-disk profile, not an empty store.
listing=$("$AIM_TEST_BINARY" list "$AIM_AGENT")
printf '%s\n' "$listing" | grep -F "$AIM_TEST_EXPECTED_STORE/profiles/$AIM_TEST_EXPECTED_PROFILE" >/dev/null
# Reset an inherited value so only reloading the parent's configuration can pass.
if [ "$AIM_PROFILE" = overridden ]; then
  env AIM_TEST_INHERITED=nested-stale "$AIM_TEST_BINARY" shell "$AIM_AGENT" overridden <<'NESTED'
set -eu
[ "$AIM_TEST_INHERITED" = profile-value ]
[ "$AIM_SESSION_ID" = explicit-session ]
NESTED
fi
echo 'PASS: environment, overrides, and nested AIM profile lookup'
CHECK
  done

  status=0
  env HOME="$AIM_REAL_HOME" "$AIM_TEST_BINARY" shell "$agent" plain <<'CHECK' || status=$?
exit 17
CHECK
  [ "$status" -eq 17 ] || { echo "FAIL: $agent shell returned $status, expected 17"; exit 1; }
  echo "PASS: $agent shell exit code propagated"
done
done
echo 'ALL LIVE CLI SHELL TESTS PASSED'
