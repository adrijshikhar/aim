#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# Upstream Canary Watcher for AIM
# Proactively tests AIM compatibility against latest upstream AI agent CLI releases
# (Codex, Claude Code, Antigravity, OpenCode).
#
# If breaking changes or runtime regressions are detected, a GitHub issue is
# filed automatically to catch regressions before end users experience them.
# ==============================================================================

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET_AGENT="all"
TARGET_VERSION="latest"
FORCE_RUN=false
DRY_RUN=false
OUTPUT_JSON=""
GITHUB_REPO="${GITHUB_REPOSITORY:-aim-cli/aim}"

print_usage() {
  cat << EOF
Usage: $0 [OPTIONS]

Options:
  --agent <name>       Target agent: 'codex', 'claude', 'agy', 'opencode', or 'all' (default: all)
  --version <ver>      Specific version to test, or 'latest' (default: latest)
  --force              Run canary checks even if local version already matches upstream
  --dry-run            Run diagnostics but do not create GitHub issues on failure
  --json <file>        Write JSON report to specified file
  --repo <owner/repo>  GitHub repository for issue filing (default: ${GITHUB_REPO})
  -h, --help           Show this help message
EOF
}

# Parse CLI arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --agent)
      TARGET_AGENT="$2"
      shift 2
      ;;
    --version)
      TARGET_VERSION="$2"
      shift 2
      ;;
    --force)
      FORCE_RUN=true
      shift
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    --json)
      OUTPUT_JSON="$2"
      shift 2
      ;;
    --repo)
      GITHUB_REPO="$2"
      shift 2
      ;;
    -h|--help)
      print_usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      print_usage
      exit 1
      ;;
  esac
done

# Color styling
BOLD="\033[1m"
GREEN="\033[32m"
YELLOW="\033[33m"
RED="\033[31m"
CYAN="\033[36m"
DIM="\033[2m"
RESET="\033[0m"

log_info()  { echo -e "${CYAN}${BOLD}[canary]${RESET} $*"; }
log_ok()    { echo -e "${GREEN}${BOLD}[ok]${RESET} $*"; }
log_warn()  { echo -e "${YELLOW}${BOLD}[warn]${RESET} $*"; }
log_err()   { echo -e "${RED}${BOLD}[fail]${RESET} $*"; }

# Temporary isolated sandbox
SANDBOX_DIR=$(mktemp -d)
trap 'rm -rf "$SANDBOX_DIR"' EXIT

export AIM_HOME="$SANDBOX_DIR/aim_home"
export AIM_REAL_HOME="$SANDBOX_DIR/real_home"
mkdir -p "$AIM_REAL_HOME/.ssh"
touch "$AIM_REAL_HOME/.gitconfig"
touch "$AIM_REAL_HOME/.fish-personal"
mkdir -p "$AIM_REAL_HOME/.local/share/omf"
mkdir -p "$AIM_REAL_HOME/.local/share/fish"
mkdir -p "$AIM_REAL_HOME/.cargo"
touch "$AIM_REAL_HOME/.cargo/env"
touch "$AIM_REAL_HOME/.cargo/env.fish"

NPM_DIR="$SANDBOX_DIR/npm"
mkdir -p "$NPM_DIR"
export PATH="$NPM_DIR/bin:$NPM_DIR/node_modules/.bin:$PATH"

# Build AIM binary
AIM_BIN="$SANDBOX_DIR/aim"
log_info "Compiling fresh AIM binary from $REPO_ROOT..."
(cd "$REPO_ROOT" && go build -o "$AIM_BIN" ./cmd/aim)

# Agent Package / Binary Mapping
get_agent_npm_pkg() {
  case "$1" in
    codex)    echo "@openai/codex" ;;
    claude)   echo "@anthropic-ai/claude-code" ;;
    opencode) echo "opencode-ai" ;;
    agy)      echo "@google/antigravity" ;;
    *)        echo "" ;;
  esac
}

get_agent_bin_name() {
  case "$1" in
    codex)    echo "codex" ;;
    claude)   echo "claude" ;;
    opencode) echo "opencode" ;;
    agy)      echo "agy" ;;
    *)        echo "$1" ;;
  esac
}

# Resolve latest upstream release version from npm
resolve_latest_version() {
  local agent="$1"
  local pkg
  pkg=$(get_agent_npm_pkg "$agent")
  if [ -n "$pkg" ]; then
    npm view "$pkg" version 2>/dev/null || echo "unknown"
  else
    echo "unknown"
  fi
}

# Check if an issue already exists on GitHub
find_existing_issue() {
  local agent="$1"
  local version="$2"
  if ! command -v gh >/dev/null 2>&1 || [ -z "${GH_TOKEN:-${GITHUB_TOKEN:-}}" ]; then
    return 1
  fi
  local search_query="in:title [Canary Alert] Upstream breaking change detected in ${agent} (${version})"
  gh issue list -R "$GITHUB_REPO" --search "$search_query" --state open --json number,url --jq '.[0].url // empty' 2>/dev/null || true
}

# Proactively file an issue on GitHub
file_canary_issue() {
  local agent="$1"
  local version="$2"
  local stage="$3"
  local details="$4"
  local issue_file="$SANDBOX_DIR/issue_body.md"

  local existing_url
  existing_url=$(find_existing_issue "$agent" "$version")
  if [ -n "$existing_url" ]; then
    log_warn "Active issue already open for $agent ($version): $existing_url"
    return 0
  fi

  cat << EOF > "$issue_file"
### 🚨 Upstream Canary Watcher Alert

An upstream breaking change or regression was detected during automated canary testing.

- **Target Agent**: \`${agent}\`
- **Upstream Version**: \`${version}\`
- **Failing Stage**: \`${stage}\`
- **Runner Environment**: \`$(uname -s) $(uname -m)\`
- **Timestamp**: \`$(date -u +"%Y-%m-%d %H:%M:%SZ")\`

---

#### Error & Diagnostic Output
\`\`\`
${details}
\`\`\`

---

#### Recommended Actions
1. Inspect upstream release notes and diffs for breaking flag changes, directory layout changes, or missing binary packaging.
2. Update the adapter in \`internal/agents/${agent}/\` to handle the new upstream behavior or flags.
3. Add a regression test to \`test/e2e_parameterized_test.sh\` verifying the fix.

*Automated report generated by \`scripts/upstream_canary.sh\`.*
EOF

  if [ "$DRY_RUN" = "true" ]; then
    log_warn "[dry-run] Skipping GitHub issue creation. Issue body written to: $issue_file"
    cat "$issue_file"
    return 0
  fi

  if command -v gh >/dev/null 2>&1 && [ -n "${GH_TOKEN:-${GITHUB_TOKEN:-}}" ]; then
    log_info "Filing GitHub issue in $GITHUB_REPO..."
    local issue_url
    issue_url=$(gh issue create -R "$GITHUB_REPO" \
      --title "[Canary Alert] Upstream breaking change detected in ${agent} (${version})" \
      --label "bug,canary,upstream-drift" \
      --body-file "$issue_file")
    log_ok "Issue filed successfully: $issue_url"
  else
    log_warn "GitHub CLI ('gh') or GH_TOKEN not available; issue not filed to remote."
  fi
}

TEST_RESULTS=()

test_agent_canary() {
  local agent="$1"
  local target_ver="$TARGET_VERSION"
  local pkg
  pkg=$(get_agent_npm_pkg "$agent")
  local bin
  bin=$(get_agent_bin_name "$agent")

  log_info "=========================================================="
  log_info "Checking Upstream Canary for Agent: ${BOLD}${agent}${RESET}"

  if [ "$target_ver" = "latest" ]; then
    log_info "Querying latest upstream version for $pkg..."
    target_ver=$(resolve_latest_version "$agent")
  fi

  if [ "$target_ver" = "unknown" ] || [ -z "$target_ver" ]; then
    log_warn "Could not resolve upstream package version for $agent (skipped or private)."
    return 0
  fi

  log_info "Target upstream version: ${BOLD}${target_ver}${RESET}"

  # Step 1: Install upstream agent in isolated sandbox
  log_info "Step 1: Installing $pkg@$target_ver in sandbox..."
  local install_out
  if ! install_out=$(npm install --prefix "$NPM_DIR" "$pkg@$target_ver" 2>&1); then
    log_err "Failed to install $pkg@$target_ver via npm"
    file_canary_issue "$agent" "$target_ver" "NPM_INSTALL" "$install_out"
    TEST_RESULTS+=("{\"agent\":\"$agent\",\"version\":\"$target_ver\",\"status\":\"FAIL\",\"stage\":\"NPM_INSTALL\"}")
    return 1
  fi
  log_ok "Installed $pkg@$target_ver"

  # Step 2: Native Binary Execution Sanity Check
  log_info "Step 2: Checking native CLI invocation ($bin --version)..."
  local ver_out
  if ! ver_out=$("$bin" --version 2>&1); then
    log_err "Native CLI invocation failed for $bin ($target_ver): $ver_out"
    file_canary_issue "$agent" "$target_ver" "UPSTREAM_BIN_SANITY" "$ver_out"
    TEST_RESULTS+=("{\"agent\":\"$agent\",\"version\":\"$target_ver\",\"status\":\"FAIL\",\"stage\":\"UPSTREAM_BIN_SANITY\"}")
    return 1
  fi
  log_ok "Upstream CLI responsive: $ver_out"

  # Step 3: AIM Profile Provisioning & Dotfile Bridging
  log_info "Step 3: Provisioning AIM profile and dotfile bridging..."
  local prof_name="canary-prof"
  export AIM_AUTO_CREATE=1
  local prof_dir="$AIM_HOME/profiles/$prof_name"
  mkdir -p "$prof_dir"

  # Seed mock credentials so profile has valid credentials
  case "$agent" in
    codex)
      mkdir -p "$prof_dir/.codex"
      echo '{"tokens":{"access_token":"canary-test"}}' > "$prof_dir/.codex/auth.json"
      ;;
    claude)
      mkdir -p "$prof_dir/.claude"
      echo '{"sessionKey":"canary-test"}' > "$prof_dir/.claude/auth.json"
      ;;
    agy)
      mkdir -p "$prof_dir/.gemini/antigravity-cli"
      echo '{"access_token":"canary-test"}' > "$prof_dir/.gemini/antigravity-cli/token.json"
      ;;
  esac

  # Populate profile dotfiles via doctor
  "$AIM_BIN" doctor "$agent" >/dev/null 2>&1 || true

  # Step 4: AIM Doctor Diagnostics Check
  log_info "Step 4: Running AIM Doctor diagnostics (--check)..."
  local doc_out
  if ! doc_out=$("$AIM_BIN" doctor "$agent" --check 2>&1); then
    log_err "AIM doctor reported failure for $agent ($target_ver)"
    echo "$doc_out"
    file_canary_issue "$agent" "$target_ver" "AIM_DOCTOR_CHECK" "$doc_out"
    TEST_RESULTS+=("{\"agent\":\"$agent\",\"version\":\"$target_ver\",\"status\":\"FAIL\",\"stage\":\"AIM_DOCTOR_CHECK\"}")
    return 1
  fi
  log_ok "AIM Doctor diagnostics passed for $agent"

  # Step 5: AIM Flag Forwarding & Execution Sanity
  log_info "Step 5: Verifying flag forwarding via AIM runner..."
  local run_out
  if ! run_out=$("$AIM_BIN" run "$agent" "$prof_name" --version 2>&1); then
    log_err "AIM runner failed to execute $agent with forwarded flags: $run_out"
    file_canary_issue "$agent" "$target_ver" "AIM_RUNNER_FORWARDING" "$run_out"
    TEST_RESULTS+=("{\"agent\":\"$agent\",\"version\":\"$target_ver\",\"status\":\"FAIL\",\"stage\":\"AIM_RUNNER_FORWARDING\"}")
    return 1
  fi
  log_ok "AIM runner successfully forwarded flags and executed $agent: $run_out"

  # Step 6: AIM Sessions Explorer Sanity
  log_info "Step 6: Verifying AIM sessions explorer compatibility..."
  local sess_out
  if ! sess_out=$("$AIM_BIN" sessions "$agent" 2>&1); then
    log_err "AIM sessions failed for $agent: $sess_out"
    file_canary_issue "$agent" "$target_ver" "AIM_SESSIONS" "$sess_out"
    TEST_RESULTS+=("{\"agent\":\"$agent\",\"version\":\"$target_ver\",\"status\":\"FAIL\",\"stage\":\"AIM_SESSIONS\"}")
    return 1
  fi
  log_ok "AIM sessions listing responsive"

  TEST_RESULTS+=("{\"agent\":\"$agent\",\"version\":\"$target_ver\",\"status\":\"PASS\",\"stage\":\"ALL\"}")
  log_ok "All canary tests passed cleanly for ${BOLD}${agent}${RESET} (${target_ver})!"
  return 0
}

OVERALL_SUCCESS=true

if [ "$TARGET_AGENT" = "all" ]; then
  AGENTS=("codex" "claude")
else
  AGENTS=("$TARGET_AGENT")
fi

for AG in "${AGENTS[@]}"; do
  if ! test_agent_canary "$AG"; then
    OVERALL_SUCCESS=false
  fi
done

# Write JSON report if requested
if [ -n "$OUTPUT_JSON" ]; then
  mkdir -p "$(dirname "$OUTPUT_JSON")"
  echo "[" > "$OUTPUT_JSON"
  FIRST=true
  for RES in "${TEST_RESULTS[@]}"; do
    if [ "$FIRST" = true ]; then
      echo "  $RES" >> "$OUTPUT_JSON"
      FIRST=false
    else
      echo "  ,$RES" >> "$OUTPUT_JSON"
    fi
  done
  echo "]" >> "$OUTPUT_JSON"
  log_info "JSON report written to $OUTPUT_JSON"
fi

if [ "$OVERALL_SUCCESS" = true ]; then
  log_ok "=========================================================="
  log_ok "All upstream agent canary checks PASSED!"
  exit 0
else
  log_err "=========================================================="
  log_err "One or more upstream agent canary checks FAILED."
  exit 1
fi
