#!/bin/sh
# AIM (AI Multiplexer) Standalone Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/adrijshikhar/aim/main/install.sh | sh

set -eu

# Formatting helpers
BOLD="$(tput bold 2>/dev/null || echo '')"
GREEN="$(tput setaf 2 2>/dev/null || echo '')"
YELLOW="$(tput setaf 3 2>/dev/null || echo '')"
BLUE="$(tput setaf 4 2>/dev/null || echo '')"
RED="$(tput setaf 1 2>/dev/null || echo '')"
RESET="$(tput sgr0 2>/dev/null || echo '')"

info() {
  printf "${BLUE}==>${RESET} ${BOLD}%s${RESET}\n" "$1"
}

success() {
  printf "${GREEN}✓${RESET} %s\n" "$1"
}

warn() {
  printf "${YELLOW}!${RESET} %s\n" "$1"
}

error() {
  printf "${RED}✗ Error:${RESET} %s\n" "$1" >&2
  exit 1
}

# 1. Detect OS
OS="$(uname -s)"
case "$OS" in
  Darwin*) OS="darwin" ;;
  Linux*)  OS="linux" ;;
  *)       error "Unsupported operating system: $OS. AIM currently supports macOS (darwin) and Linux." ;;
esac

# 2. Detect Architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)   ARCH="amd64" ;;
  arm64|aarch64)  ARCH="arm64" ;;
  *)              error "Unsupported architecture: $ARCH. AIM supports amd64 and arm64." ;;
esac

# 3. Resolve Target Version
REPO="adrijshikhar/aim"
if [ -n "${AIM_VERSION:-}" ]; then
  VERSION="$AIM_VERSION"
else
  info "Checking latest release of AIM..."
  LATEST_JSON="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null || echo '')"
  if [ -n "$LATEST_JSON" ]; then
    VERSION="$(echo "$LATEST_JSON" | grep '"tag_name":' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/' || echo '')"
  fi

  if [ -z "${VERSION:-}" ]; then
    # Fallback to redirect URL lookup if API rate-limited
    REDIRECT_URL="$(curl -fsSLI -o /dev/null -w "%{url_effective}" "https://github.com/${REPO}/releases/latest" 2>/dev/null || echo '')"
    VERSION="$(basename "$REDIRECT_URL")"
  fi

  if [ -z "${VERSION:-}" ] || [ "$VERSION" = "latest" ]; then
    VERSION="v0.1.0"
  fi
fi

# Clean version tag (strip leading v for archive filename)
CLEAN_VERSION="${VERSION#v}"
TARBALL_NAME="aim_${CLEAN_VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL_NAME}"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

info "Installing AIM ${VERSION} (${OS}/${ARCH})..."

# 4. Prepare Temporary Directory
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

# 5. Download Artifacts
info "Downloading ${TARBALL_NAME}..."
DOWNLOAD_SUCCESS=false

if curl -fSL --progress-bar "$DOWNLOAD_URL" -o "$TMP_DIR/$TARBALL_NAME" 2>/dev/null; then
  DOWNLOAD_SUCCESS=true
elif [ -n "${GITHUB_TOKEN:-}" ] || [ -n "${GH_TOKEN:-}" ]; then
  TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
  if curl -fSL -H "Authorization: token $TOKEN" "$DOWNLOAD_URL" -o "$TMP_DIR/$TARBALL_NAME" 2>/dev/null; then
    DOWNLOAD_SUCCESS=true
  fi
fi

if [ "$DOWNLOAD_SUCCESS" = "false" ] && command -v gh >/dev/null 2>&1; then
  if gh release download "$VERSION" -R "$REPO" -p "$TARBALL_NAME" -D "$TMP_DIR" >/dev/null 2>&1; then
    DOWNLOAD_SUCCESS=true
    gh release download "$VERSION" -R "$REPO" -p "checksums.txt" -D "$TMP_DIR" >/dev/null 2>&1 || true
  fi
fi

if [ "$DOWNLOAD_SUCCESS" = "false" ]; then
  error "Failed to download $DOWNLOAD_URL. Please verify your internet connection, release visibility, or GitHub token."
fi

# 6. Verify Checksum
if [ -f "$TMP_DIR/checksums.txt" ] || curl -fsSL "$CHECKSUMS_URL" -o "$TMP_DIR/checksums.txt" 2>/dev/null; then
  info "Verifying SHA256 checksum..."
  EXPECTED_SUM="$(grep "$TARBALL_NAME" "$TMP_DIR/checksums.txt" | awk '{print $1}' || echo '')"
  if [ -n "$EXPECTED_SUM" ]; then
    if command -v shasum >/dev/null 2>&1; then
      ACTUAL_SUM="$(shasum -a 256 "$TMP_DIR/$TARBALL_NAME" | awk '{print $1}')"
    elif command -v sha256sum >/dev/null 2>&1; then
      ACTUAL_SUM="$(sha256sum "$TMP_DIR/$TARBALL_NAME" | awk '{print $1}')"
    else
      ACTUAL_SUM=""
    fi

    if [ -n "$ACTUAL_SUM" ]; then
      if [ "$EXPECTED_SUM" != "$ACTUAL_SUM" ]; then
        error "Checksum verification failed! Expected: $EXPECTED_SUM, Got: $ACTUAL_SUM"
      fi
      success "Checksum verified"
    fi
  fi
fi

# 7. Extract Binary
tar -xzf "$TMP_DIR/$TARBALL_NAME" -C "$TMP_DIR"
if [ ! -f "$TMP_DIR/aim" ]; then
  error "Failed to locate 'aim' binary in extracted archive."
fi

# 8. Determine Installation Directory
if [ -n "${AIM_INSTALL_DIR:-}" ]; then
  DEST_DIR="$AIM_INSTALL_DIR"
elif [ -w "/usr/local/bin" ]; then
  DEST_DIR="/usr/local/bin"
else
  DEST_DIR="${HOME}/.local/bin"
fi

mkdir -p "$DEST_DIR"
cp "$TMP_DIR/aim" "$DEST_DIR/aim"
chmod 755 "$DEST_DIR/aim"
success "Installed aim to ${DEST_DIR}/aim"

# 9. Install Shell Completions (best effort)
if command -v "$DEST_DIR/aim" >/dev/null 2>&1; then
  # zsh
  ZSH_COMP_DIR="${HOME}/.zsh/completions"
  if mkdir -p "$ZSH_COMP_DIR" 2>/dev/null; then
    "$DEST_DIR/aim" completion zsh > "$ZSH_COMP_DIR/_aim" 2>/dev/null || true
  fi

  # bash
  BASH_COMP_DIR="${HOME}/.local/share/bash-completion/completions"
  if mkdir -p "$BASH_COMP_DIR" 2>/dev/null; then
    "$DEST_DIR/aim" completion bash > "$BASH_COMP_DIR/aim" 2>/dev/null || true
  fi

  # fish
  FISH_COMP_DIR="${HOME}/.config/fish/completions"
  if mkdir -p "$FISH_COMP_DIR" 2>/dev/null; then
    "$DEST_DIR/aim" completion fish > "$FISH_COMP_DIR/aim.fish" 2>/dev/null || true
  fi
  success "Shell completions installed for zsh, bash, fish"
fi

# 10. Check PATH
case ":$PATH:" in
  *":$DEST_DIR:"*) ;;
  *)
    warn "${DEST_DIR} is not in your PATH."
    printf "  Add it to your shell configuration:\n"
    printf "    ${BOLD}export PATH=\"%s:\$PATH\"${RESET}\n\n" "$DEST_DIR"
    ;;
esac

printf "\n${GREEN}${BOLD}AIM (AI Multiplexer) successfully installed!${RESET}\n"
printf "Run ${BOLD}aim${RESET} to open the interactive TUI, or ${BOLD}aim --help${RESET} for commands.\n"
