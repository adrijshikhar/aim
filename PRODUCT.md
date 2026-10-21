# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users
Software engineers, AI agent power users, and development team leads managing multiple AI coding assistants (Antigravity/agy, Codex, Claude Code, Gemini). They switch between multiple accounts (work, personal, client, enterprise) throughout the day, navigate strict rate limits and quota resets, and collaborate with teammates over local network or VPN meshes to pool quota.

## Product Purpose
AIM (AI Multiplexer) eliminates account collision, quota exhaustion, and state contamination across AI coding assistants. It provides isolated virtual homes for each profile, real-time quota telemetry with automated background pre-warming, peer-to-peer credential leasing/quota pooling, and seamless session resumption across both terminal and web interfaces. Success means zero credential leakage, instant quota visibility (<1ms cached), and fluid, unified control whether operating directly in the terminal (CLI/TUI) or via the interactive Web Dashboard.

## Positioning
Unlike generic profile switchers or shell wrappers, AIM provides sandboxed virtual home directories (`~/.aim/profiles/<name>`) that dynamically bridge host dotfiles (`.gitconfig`, `.ssh`) and MCP configurations while guaranteeing zero credential pollution. Furthermore, its P2P mesh network allows developers on the same Wi-Fi or internal VPN (Pritunl/OpenVPN) to expose and borrow profile quota leases in real-time under isolated borrowed scopes (`host:<profile>`) without exchanging persistent secrets.

## Operating Context
- Terminal environments: developers running headless CLI commands (`aim run`, `aim usage`, `aim doctor`, `aim list`) and interactive Bubble Tea TUI dashboards (`aim`).
- Browser environments: interactive React 19 Web Dashboard (`http://localhost:18480`) designed as an interactive, graphical way to use the AIM CLI and manage network/mesh operations.
- Network environments: local developer workstations, internal VPNs, and LAN meshes sharing profile quota leases via a dedicated sharing daemon (port 4319).
- Background services: native OS daemon integration (`launchd` on macOS, `systemd` on Linux) performing automated 15-minute quota cache pre-warming.

## Capabilities and Constraints
- Profile Isolation: Virtual home directories per profile; isolated auth, credentials, and conversation histories.
- Agent Support: Antigravity (`agy`), Gemini (`gemini`), Claude Code (`claude`), Codex (`codex`).
- Dual Interface: Interactive Bubble Tea TUI / Lipgloss CLI alongside a React 19 + TypeScript + Vite + Tailwind CSS Web Dashboard. The dashboard's primary role is to provide an interactive way to use the AIM CLI and manage network/mesh operations.
- P2P Quota Pooling: Profile sharing over internal mesh with auto-approve or manual lease requests, TTL leases, and borrow isolation.
- Security Invariants: Directory permissions `0700`, credential permissions `0600`, zero secret leakage in logs/UI, atomic file replacements.
- Platform Non-Interference: In-place repo execution (`./aim`); strictly prohibited from modifying user shell profiles, system PATHs, or touching `/opt/homebrew/bin` or `~/.local/bin`.
- Concurrency & Reliability: Zero data races (`go test -race ./...`), context cancellation propagation, no leaked goroutines.

## Brand Commitments
- Theme & Palette: Strict Atom One Dark semantic token hierarchy (`internal/tui/styles.go`), emphasizing low visual fatigue, chalky tones, and high information density.
- Voice & Tone: Developer-first, minimalist, authoritative, and precise. No fluff, no marketing hype.
- Terminal Fidelity: Bounded tables (`presenter.NewTable`), strict ANSI-aware Lipgloss styling, zero broken cell wrapping or ad-hoc math.

## Evidence on Hand
- Working Go CLI and Bubble Tea TUI with comprehensive test suites.
- Production React 19 web application in `web/` with Radix UI, Lucide icons, and Tailwind CSS.
- Atom One Dark design specification in `DESIGN.md` and `internal/tui/styles.go`.
- Real-time P2P sharing mesh protocol implemented in Go with HTTP REST endpoints and WebSocket/event streaming.
- Comprehensive documentation in `README.md`, `wiki.md`, `TROUBLESHOOTING.md`, and `AGENTS.md`.

## Product Principles
1. **Developer Rigor & Low Fatigue**: Prioritize speed, high information density, and the Atom One Dark aesthetic across both terminal and web surfaces.
2. **Strict Boundary Isolation**: Keep agent states, OAuth tokens, and profile directories strictly isolated; never risk credential leakage or shell pollution.
3. **Dual Surface Cohesion**: The Web Dashboard and CLI/TUI are harmonious complements—the web interface serves as an intuitive, visual vehicle for the CLI's capabilities without deviating from core CLI semantics.
4. **Bottleneck-First Telemetry**: Highlight quota capacity, reset timers, and active constraints immediately to prevent workflow interruptions.
5. **Team Quota Liquidity**: Make quota sharing across teammates effortless and safe over trusted networks.

## Accessibility & Inclusion
- High-contrast text tokens (`TextBright: #e5e5e5`) for active selections and keycaps.
- Multi-tier color indicators backed by text labels (e.g. `[OK]`, `[WARN]`, `[FAIL]`) to ensure usability under color-blindness and `NO_COLOR` environments.
- Fully accessible keyboard navigation in both Bubble Tea TUI and Radix-based web components.
