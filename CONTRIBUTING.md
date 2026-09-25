# Contributing to AIM

Thank you for your interest in contributing to AIM (AI Multiplexer)! 

AIM is built in Go to provide lightweight, rock-solid profile isolation and quota tracking for AI coding agents like Antigravity, Gemini, Claude Code, and Codex.

---

## 1. Development Setup

### Prerequisites
* **Go**: `1.23` or later
* **Git**: `2.30` or later
* **Make**: standard Unix make
* **Optional**: [GoReleaser](https://goreleaser.com) (`brew install goreleaser`) for release packaging

### Clone & Build
```bash
git clone https://github.com/adrijshikhar/aim.git
cd aim

# Build local binary to ./aim
make build

# Install locally to ~/.local/bin/aim (with shell completions)
make install
```

---

## 2. Testing & Quality Gates

AIM enforces strict testing standards including race detection and integration smoke testing.

```bash
# Run unit tests with the race detector
make test

# Run end-to-end integration smoke tests
make smoke

# Run the built AIM CLI through real shells and nested AIM commands
make live-test

# Format code
gofmt -s -w .

# Static analysis
go vet ./...
```

`make smoke` uses mocked agent binaries. `make live-test` launches real `/bin/sh`
processes through each adapter and executes nested AIM commands against temporary
profiles. It checks profile-store continuity, environment isolation, configuration
overrides, and exit codes in custom `AIM_HOME`, legacy `.aim`, custom XDG, and
platform-default layouts; only macOS Keychain access is stubbed. Neither suite
authenticates with providers or verifies a live model response.

Profile children carry resolved AIM paths in `AIM_CONFIG_DIR`, `AIM_DATA_DIR`,
`AIM_CACHE_DIR`, and `AIM_STATE_DIR`, plus `AIM_REAL_HOME`. These preserve nested
AIM commands without changing the provider's global `XDG_*` environment.
`AIM_HOME` still selects the legacy layout and takes precedence over these paths.

### Local Release Testing
To verify cross-compilation across all target platforms (`darwin/arm64`, `darwin/amd64`, `linux/amd64`, `linux/arm64`):
```bash
make release-snapshot
```

---

## 3. Adding a New Agent Adapter

AIM is designed to be easily extensible to new AI agents. Adding support for an agent requires implementing the `AgentAdapter` interface.

### Step 1: Create the Adapter
> [!TIP]
> For complete architectural guidelines, keychain mitigation rules, hook trust path rewriting, and sidecar proxy handling when building adapters, consult the comprehensive onboarding guide at [`skills/onboard-provider/SKILL.md`](skills/onboard-provider/SKILL.md).

Create a new package under `internal/agents/<name>/adapter.go`:

```go
package myagent

import (
	"context"
	"github.com/aim-cli/aim/internal/agents"
	"github.com/aim-cli/aim/internal/usage"
)

type Adapter struct{}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Name() string        { return "myagent" }
func (a *Adapter) DisplayName() string { return "My Agent CLI" }
func (a *Adapter) Aliases() []string   { return []string{"ma"} }
func (a *Adapter) BinaryName() string  { return "myagent" }

func (a *Adapter) HasCredentials(profileDir string) bool {
	// Return true if the profile directory contains valid tokens/credentials
	return true
}

func (a *Adapter) Login(ctx context.Context, profileName, profileDir string) error {
	// Execute OAuth PKCE or spawn agent interactive login
	return nil
}

func (a *Adapter) PrepareEnv(profileName, profileDir string) (agents.LaunchEnv, error) {
	// Configure environment variables (e.g. HOME=profileDir) and binary path
	return agents.LaunchEnv{
		BinaryPath: a.BinaryName(),
		Env: map[string]string{
			"HOME": profileDir,
		},
	}, nil
}

func (a *Adapter) Doctor(ctx context.Context, profileName, profileDir string) []agents.DiagnosticResult {
	// Return health check diagnostics (binary presence, token expiry, config status)
	return nil
}

func (a *Adapter) GetUsage(ctx context.Context, profileName, profileDir string) (*usage.Report, error) {
	// Return quota metrics and reset countdowns (or return nil if unsupported)
	return nil, nil
}
```

### Step 2: Register in `cmd/aim/main.go`
Register the new adapter in the central agent registry:

```go
reg.Register(myagent.NewAdapter())
```

Once registered, the new agent will automatically be supported across:
* **CLI execution**: `aim run myagent <profile>`, `aim shell myagent <profile>`, `aim login myagent <profile>`
* **Shell autocompletions**: `zsh`, `bash`, and `fish` tab-completion
* **TUI Dashboard**: Interactive tabs and profile filtering
* **Diagnostics**: `aim doctor myagent`

---

## 4. Developer Tool & Dotfile Bridging Guidelines

When extending bridged configurations in `internal/profile/symlink.go`:
1. **Never bridge agent state**: Do not bridge directories containing AI agent credentials or history (`.aim`, `.gemini`, `.claude`, `.codex`) to maintain profile isolation.
2. **Validate Path Locality**: Always sanitize custom paths with `filepath.IsLocal` to prevent directory traversal attacks.
3. **Non-Destructive Linking**: Existing files or custom configurations in a profile directory must never be overwritten.

---

## 5. Pull Request Process

1. Fork the repository and create your branch from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```
2. Make your changes with accompanying unit tests in `*_test.go`.
3. Ensure all tests pass:
   ```bash
   make test && make smoke
   ```
4. Commit using [Conventional Commits](https://www.conventionalcommits.org/):
   * `feat: add support for Claude Code adapter`
   * `fix: handle edge case in usage cache invalidation`
   * `docs: update installation instructions`
5. Open a Pull Request with a clear description of the changes and any testing performed.
