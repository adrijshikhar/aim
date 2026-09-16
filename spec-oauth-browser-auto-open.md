# Technical Design Specification: Browser Auto-Open on OAuth Expiry & Re-Authentication

- **Author**: Antigravity Engineering
- **Date**: 2026-09-16
- **Status**: Draft / In-Review
- **Target Subsystems**: `internal/agents/agy`, `internal/runner`, `cmd/aim`

---

## 1. Problem Statement & Motivation

When an Antigravity (`agy`) OAuth access token expires and its refresh token is revoked or cannot be refreshed (leaving the profile in an `[offline]` or unauthenticated state), launching Antigravity via `aim` or inside a subshell results in a broken login experience:

1. Antigravity prompts the user to authenticate (`Login`).
2. Antigravity starts an OAuth PKCE loopback flow.
3. **Bug**: Antigravity does **not** launch Google Chrome or the default web browser.
4. The user is forced to manually highlight, copy, and paste the authorization URL from the terminal into their browser.

### Root Cause
1. In `internal/agents/agy/adapter.go` (`PrepareEnv`):
   ```go
   if a.HasCredentials(profileDir) {
       envMap["SSH_CONNECTION"] = "127.0.0.1 50000 127.0.0.1 22"
   } else {
       delete(envMap, "SSH_CONNECTION")
   }
   ```
2. `HasCredentials(profileDir)` only checks if `antigravity-oauth-token` exists and is non-empty (`len(data) > 0`). It does **not** check whether the token is valid, expired, or rejected.
3. When credentials expire, the file remains on disk. `HasCredentials` evaluates to `true`.
4. `PrepareEnv` injects `SSH_CONNECTION="127.0.0.1 50000 127.0.0.1 22"`.
5. Inside the Antigravity binary (`/Users/nemesis/.local/bin/agy`), the Code Assist authentication engine (`keyring_detector_ssh.go` and `/auth/callback`) checks for `SSH_CONNECTION`.
6. When `SSH_CONNECTION` is detected, Antigravity assumes the process is running inside a remote headless SSH terminal. It **deliberately suppresses browser auto-open** (`exec.Command("open", url)`) and falls back to console-only URL printing.
7. Furthermore, in `internal/runner/exec.go`, the background Keychain harvester and auto-purger are guarded by `if !hasCreds`. Because the expired file exists on disk, `hasCreds` evaluates to `true`, preventing newly authenticated Keychain tokens from being automatically harvested back into the profile.

---

## 2. Goals & Non-Goals

### Goals
- **Seamless Browser Redirection**: When a profile's credentials are missing, expired, or invalid, launching Antigravity or clicking "Login" must automatically open the user's default browser (Google Chrome) with the OAuth consent URL.
- **Strict Profile Isolation**: Ensure that unauthenticated and re-authenticating sessions do not leak tokens into the global macOS Keychain or contaminate other profiles.
- **Automatic Keychain Harvesting on Re-Auth**: If a session transitions from unauthenticated/expired to authenticated while `agy` runs, the new token must be immediately captured to the profile directory and scrubbed from macOS Keychain.
- **Deterministic Token Health Evaluation**: Differentiate between:
  1. *Healthy / Active* (Valid access token or valid refresh token): Keep keyring bypass active (`SSH_CONNECTION`).
  2. *Expired / Offline / Unauthenticated*: Omit `SSH_CONNECTION` so browser auto-launches, and activate Keychain harvester.

### Non-Goals
- Changing Antigravity's internal binary behavior (we must operate via external environment variables and process supervisors).
- Interfering with other agents (`codex`, `gemini`, `claude`) which have independent credential managers.

---

## 3. Architecture & Technical Design

### 3.1 Token State Evaluation (`IsTokenHealthy`)

Add a helper to `internal/agents/agy/adapter.go`:

```go
// TokenHealth represents the authentication state of an Antigravity token file.
type TokenHealth int

const (
    TokenHealthMissing TokenHealth = iota
    TokenHealthInvalid
    TokenHealthExpired
    TokenHealthValid
)

func (a *Adapter) CheckTokenHealth(profileDir string) TokenHealth
```

Criteria:
1. **Missing**: File does not exist or size == 0.
2. **Invalid**: File exists but contains malformed JSON or lacks both `access_token` and `refresh_token`.
3. **Expired**: Access token is past its `expiry` timestamp AND `refresh_token` is either absent or known to fail refresh (e.g. usage cache reports error/offline).
4. **Valid**: Access token expiry is in the future, or valid refresh token exists and profile is not offline.

### 3.2 Environment Construction in `PrepareEnv`

In `PrepareEnv(profileName, profileDir string)`:

```go
health := a.CheckTokenHealth(profileDir)
if health == TokenHealthValid {
    // Healthy session: enable keyring bypass so agy reads strictly from profile token file
    envMap["SSH_CONNECTION"] = "127.0.0.1 50000 127.0.0.1 22"
} else {
    // Missing, expired, or offline: MUST omit SSH_CONNECTION to allow browser auto-open
    delete(envMap, "SSH_CONNECTION")
}
delete(envMap, "SSH_CLIENT")
delete(envMap, "SSH_TTY")
```

### 3.3 Runner Keychain Watcher in `internal/runner/exec.go`

In `runner.Run`:
```go
hasValidCreds := false
if profileDir != "" && agentName == "agy" {
    adapter, _ := reg.Get("agy")
    if agyAdapter, ok := adapter.(*agy.Adapter); ok {
        hasValidCreds = agyAdapter.CheckTokenHealth(profileDir) == agy.TokenHealthValid
    }
}

// If credentials are NOT valid (missing OR expired), activate the background harvester
if !hasValidCreds {
    _ = profile.PurgeIgnoredKeychains(agentName, customServices...)
    // Start background watcher to harvest token from Keychain upon browser login completion
    // ...
}
```

### 3.4 Explicit Login Command (`aim login agy <profile>`)

When the user explicitly invokes `aim login agy <profile>` or presses `l` in the TUI:
- Cleanly back up any existing stale/expired token file to `.antigravity-oauth-token.stale`.
- Launch `agy` with clean environment (no `SSH_CONNECTION`).
- Upon exit, harvest new token and purge Keychain.

---

## 4. Verification & Testing Plan

### 4.1 Automated Unit Tests
- `TestAgy_PrepareEnv_TokenHealth`:
  - Case 1: Non-existent token -> `SSH_CONNECTION` omitted.
  - Case 2: Valid unexpired token -> `SSH_CONNECTION` set.
  - Case 3: Expired token with invalid refresh token -> `SSH_CONNECTION` omitted.
  - Case 4: Malformed JSON token -> `SSH_CONNECTION` omitted.
- `TestRunner_ExpiredToken_ActivatesKeychainHarvester`:
  - Verifies that `runner.Run` starts the background harvester when a token is expired.

### 4.2 Live Integration Testing (`bot` Profile)
1. **Baseline**: Back up active token from `/Users/nemesis/.aim/profiles/bot/.gemini/antigravity-cli/antigravity-oauth-token`.
2. **Test 1 (Missing Credentials)**:
   - Remove token file from `bot`.
   - Run `./aim doctor agy bot` -> verify reports missing token.
   - Run `PrepareEnv` -> verify `SSH_CONNECTION` is NOT set.
3. **Test 2 (Expired Token Simulation)**:
   - Write an expired token (`expiry: 2020-01-01T00:00:00Z` and invalid `refresh_token`).
   - Run `PrepareEnv` -> verify `SSH_CONNECTION` is NOT set.
   - Verify `aim doctor` flags token as expired.
4. **Restoration**:
   - Restore the verified backup token to `bot`.
   - Run `./aim doctor agy bot` -> verify status is restored to `[OK]`.
