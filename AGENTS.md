# AGENTS.md — AI Multiplexer (AIM) Agent Guidelines & Invariants

> **Audience:** All autonomous and pair-programming AI coding agents working on the AIM codebase.  
> **Core Mandate:** Maintain production-grade Go systems rigor, absolute terminal UI precision, zero data races, and strict credential security.

---

## 1. Architecture & Separation of Concerns

AIM operates across three distinct user interface paradigms. Agents must strictly preserve their boundaries:

1. **Interactive Full-Screen TUI (`internal/tui/`)**:
   - **Engine:** [Charm Bubble Tea](https://github.com/charmbracelet/bubbletea) (`tea.Model`, `Init`, `Update`, `View`).
   - **Scope:** Interactive terminal dashboards, keyboard-navigated session selectors, modals, and tabbed inspectors.
   - **Behavior:** Takes over terminal with alternate screen buffers (`tea.WithAltScreen()`) and raw keyboard input.
2. **Headless & Static CLI Output (`cmd/aim/`, `internal/presenter/`)**:
   - **Engine:** [Charm Lipgloss](https://github.com/charmbracelet/lipgloss) (`lipgloss.Style`, `lipgloss/table`).
   - **Scope:** Non-interactive CLI commands writing directly to stdout/stderr (`aim list`, `aim doctor`, `aim mcp list`, `aim sessions --plain`, `aim usage`, `aim whoami`).
   - **Behavior:** Prints bounded, beautifully formatted ANSI text once, honoring terminal width, piping, and `NO_COLOR`.
3. **Web Dashboard & REST Service (`internal/web/`, `web/`)**:
   - **Backend:** Go HTTP server (`internal/web/`, `internal/service/`) with embedded static assets.
   - **Frontend:** React 19 + TypeScript + Vite + Tailwind CSS + Radix UI (`web/`).

---

## 2. Strict Lipgloss & Terminal Rendering Rules

Terminals have hard width constraints. Misaligned columns, broken borders, or wrapping cells look amateur and destroy usability. All agents must enforce the following invariants:

### 2.1 Always Use Bounded Tables (`presenter.NewTable`)
- **PROHIBITED:** Never instantiate an unconstrained Lipgloss table with raw `table.New()` without setting width.
- **MANDATORY:** Always construct CLI tables using `presenter.NewTable(w)` (or explicitly chain `.Width(presenter.TerminalWidth(w, 100)).Wrap(false)`).
- **RATIONALE:** If a table's width is not constrained to the terminal window, long paths, URLs, or commands (e.g. 280+ character Java classpaths) will produce rows wider than the screen. The terminal emulator will wrap the line mid-row, pushing border characters (`│`) onto new lines and destroying table alignment.

```go
// ❌ BROKEN: Unconstrained table blows out terminal on long strings
t := table.New().
    Headers("NAME", "COMMAND").
    Row("wiremock", "/very/long/path/to/some/dependency/wiremock.jar ...")
fmt.Println(t.Render())

// ✅ CORRECT: Constrained to detected terminal width with automatic truncation
t := presenter.NewTable(out).
    Border(lipgloss.HiddenBorder()).
    Headers("NAME", "COMMAND").
    Row("wiremock", "/very/long/path/to/some/dependency/wiremock.jar ...")
fmt.Fprintln(out, t.Render())
```

### 2.2 Rely on Lipgloss Layout — No Ad-Hoc Math
- **PROHIBITED:** Never write manual rune slicing, arbitrary cutoff math (e.g., `termWidth - 63`, `maxLen - 3`), or string truncation loops to compensate for missing table width.
- **MANDATORY:** Set `t.Wrap(false)` on tables with dynamic or unpredictable content. Lipgloss's internal column optimizer (`r.optimizedWidths()`) and cell renderer (`ansi.Truncate(cell, length, "…")`) will automatically allocate proportional widths and apply ellipsis truncation cleanly within the column border.

### 2.3 Strict Atom One Dark Theme Tokens
- **PROHIBITED:** Never hardcode raw hex colors (`lipgloss.Color("#...")`) or ANSI numeric color codes (`lipgloss.Color("240")`) in CLI commands, presenters, or UI components.
- **MANDATORY:** Always import and use semantic tokens from [`internal/tui/styles.go`](file:///Users/nemesis/Projects/my-projects/aim/internal/tui/styles.go):
  - **Headers & Cursor:** `tui.AccentBlue` (`#61afef`)
  - **High-Contrast Text / Titles / Keycaps:** `tui.TextBright` (`#e5e5e5`)
  - **Standard Body Text:** `tui.TextPrimary` (`#abb2bf`)
  - **Entity Names / Scopes / Profiles:** `tui.AccentCyan` (`#56b6c2`)
  - **Dividers & Subtle Borders:** `tui.TextDim` (`#4b5263`)
  - **Hints / Timestamps / Inactive Text:** `tui.TextMuted` (`#5c6370`)
  - **Status Healthy / OK:** `tui.StatusGreen` (`#98c379`)
  - **Status Warning / Refreshing:** `tui.StatusYellow` (`#e5c07b`)
  - **Status Critical / Exhausted / Error:** `tui.StatusRed` (`#e06c75`)

### 2.4 Never Use `fmt.Sprintf` Column Padding on Styled Strings
- **PROHIBITED:**
  ```go
  // ❌ BROKEN: ANSI escape bytes corrupt string length calculation
  badge := lipgloss.NewStyle().Foreground(tui.StatusGreen).Render("[OK]")
  fmt.Sprintf("%-10s %s", badge, message) // Output will be misaligned by 9+ spaces!
  ```
- **MANDATORY:** Use Lipgloss `.Width(...)` or Lipgloss table cells:
  ```go
  // ✅ CORRECT: Lipgloss measures printable display width, ignoring ANSI escapes
  badge := lipgloss.NewStyle().Foreground(tui.StatusGreen).Width(10).Render("[OK]")
  fmt.Sprintf("%s %s", badge, message)
  ```

---

## 3. Go Systems & Concurrency Invariants

All Go code must meet the `cc-skills-golang` standard:

### 3.1 Concurrency & Race Safety
- **Zero Data Races:** All PRs and commits must pass `go test -race ./...`.
- **Context Cancellation:** Propagate `context.Context` to all blocking network, exec, and disk operations. Always include `case <-ctx.Done(): return ctx.Err()` in `select` blocks.
- **No Leaked Goroutines:** Any background goroutine (daemons, prewarmers, watchers) must have an explicit stop channel or terminate upon context cancellation.

### 3.2 Filesystem & Credential Security
- **Strict Directory & File Permissions:**
  - Directories: `0700` (`os.MkdirAll(dir, 0700)`).
  - Config & Credential files: `0600` (`os.WriteFile(path, data, 0600)`).
- **Atomic Disk Writes:** Never overwrite sensitive state files directly. Write to `.tmp.<pid>` or `os.CreateTemp`, sync, and atomically replace via `os.Rename`.
- **Zero Credential Leakage:** Tokens, refresh secrets, and OAuth headers must never appear in log files, CLI outputs, or diagnostic reports (`aim doctor --report`). Always redact tokens before display.

### 3.3 Strict Read-Only Inspection Guarantee
- Commands like `aim list`, `aim doctor`, `aim mcp list`, `aim sessions`, and read-only API endpoints (`GET /api/profiles`) **must strictly never make network requests or rotate OAuth credentials**.
- Use [`usage.InspectProfile(...)`](file:///Users/nemesis/Projects/my-projects/aim/internal/usage/engine.go) to inspect cached or on-disk profile health without triggering token refreshes or provider rate limits.

---

## 4. Verification & Quality Gate

Before declaring any work complete or opening a PR:

1. **Format:**
   ```bash
   gofmt -s -w . && test -z "$(gofmt -l .)"
   ```
2. **Race-Free Tests:**
   ```bash
   go test -race ./...
   ```
3. **Smoke Verification:**
   Verify CLI commands produce clean, bounded output in standard terminal dimensions (80 and 120 columns).
