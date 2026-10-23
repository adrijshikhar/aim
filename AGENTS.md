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
- **PROHIBITED:** Never instantiate an unconstrained Lipgloss table with raw `table.New()` for terminal rendering without width constraints.
- **MANDATORY:** Always construct CLI tables using `presenter.NewTable(w)`. When writing to an interactive terminal, `NewTable` automatically bounds the table to the detected terminal width with `.Wrap(false)`. When writing to a pipe, subshell, or file (`!term.IsTerminal`), it leaves width unconstrained so full paths, URLs, and identifiers are preserved without ellipsis truncation (`…`).
- **RATIONALE:** If a table's width is not constrained to the terminal window, long paths, URLs, or commands (e.g. 280+ character Java classpaths) will produce rows wider than the screen. The terminal emulator will wrap the line mid-row, pushing border characters (`│`) onto new lines and destroying table alignment. Conversely, when output is piped or redirected, truncating data breaks grep, scripts, and automation.

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

### 3.4 Binary Paths & Shell Environment Isolation (STRICT NON-INTERFERENCE)
- **PROHIBITED:** Never copy, move, symlink, or install binaries into system or user binary directories (`/opt/homebrew/bin`, `~/.local/bin`, `/usr/local/bin`, `/usr/bin`, `/bin`) on local or remote machines. All build outputs must strictly remain within the repository root (e.g., `./aim`).
- **PROHIBITED:** Never modify, append to, or overwrite user shell configuration files (`~/.zshrc`, `~/.bashrc`, `~/.config/fish/config.fish`, `~/.profile`, etc.) or change system PATHs.
- **PROHIBITED:** Never touch, replace, or wrap external tool binaries, launchers, or scripts (e.g., `codex`, `claude`, `cxstatusline`, `agy`) in user directories or profile directories.
- **MANDATORY:** Test and execute binaries strictly in-place within the repository (e.g., `./aim <subcommand>`) or using isolated test fixtures. If a binary must be installed globally, ask the user or print the exact command for the user to run.

### 3.5 Safe Process Management & Prohibition of Blanket Termination (NO KILLALL)
- **PROHIBITED:** Never run `killall` (e.g., `killall aim`, `killall codex`, `killall claude`, `killall node`) or broad pattern kills (`pkill -f <pattern>`) on local or remote machines. Blanket process killers indiscriminately terminate the user's active foreground sessions, running agent jobs, and unrelated user tasks.
- **MANDATORY:** Always manage processes targeting specific, verified process IDs (PIDs). Use standard application lifecycle subcommands (e.g., `./aim web stop`), read verified PID files (e.g., `web.pid`), or target specific port listeners (`lsof -ti :<port>`). Always verify the process command line before issuing termination signals.

---

## 4. Strict Web Dashboard & React UI Invariants

All agents modifying or extending the web dashboard (`web/src/`) must strictly enforce the following frontend invariants:

### 4.1 Zero Layout Shift & Border Box-Model Parity
- **PROHIBITED:** Never toggle a 1px border on and off conditionally without maintaining `border-transparent` on the inactive state. Adding `border border-border` on active while omitting a border on inactive adds 2px to the element's box model, causing adjacent content and parent cards to jitter/jump vertically on click.
- **PROHIBITED:** Never apply `transition-all` to tabs, buttons, or segmented controls. `transition-all` interpolates height, padding, and border width during click state changes, creating visible wobbling/bouncing.
- **MANDATORY:** Always use `transition-colors duration-150` exclusively for state toggles. Always maintain border parity (`border border-transparent` on inactive vs `border border-border/80` on active) and enforce `whitespace-nowrap`.

```tsx
// ❌ BROKEN: Adds 2px border on active and interpolates dimensions with transition-all
<button className={cn("px-2.5 py-1 text-xs transition-all", isSelected ? "border border-border font-semibold" : "text-muted-foreground")}>

// ✅ CORRECT: Strict border parity and transition-colors prevent all layout shift
<button className={cn("px-2.5 py-1 text-xs border transition-colors whitespace-nowrap", isSelected ? "bg-card text-foreground border-border/80 font-semibold" : "border-transparent text-muted-foreground hover:bg-secondary/60")}>
```

### 4.2 Use Canonical UI Primitives (No Ad-Hoc Switchers)
- **PROHIBITED:** Never write ad-hoc `<button>` loops or button groups to toggle modes or options.
- **PROHIBITED:** Never use full-width grid tabs (`grid w-full grid-cols-2`) stretching across wide cards or viewports for simple 2–3 option selectors.
- **MANDATORY:** Always use the canonical [`SegmentedControl`](file:///Users/nemesis/Projects/my-projects/aim/web/src/components/ui/segmented-control.tsx) component (`size="sm"` or `size="default"`). It automatically guarantees intrinsic width, WAI-ARIA `radiogroup`/`radio` accessibility, fixed height contracts (`h-7` / `h-8`), and border parity.

### 4.3 Pure Monoline Vector Icons & Absolute Emoji Prohibition
- **PROHIBITED:** Never use raw Unicode or OS emojis (`🔒`, `✋`, `⚡`, `🚀`, `💡`, `🎉`) anywhere in the UI (dropdowns, badges, buttons, cards, table cells). Emojis render inconsistently across operating systems, break line-height alignment, and destroy the dark-tech aesthetic.
- **MANDATORY:** Use minimalist monoline vector outline icons exclusively via [`lucide-react`](https://lucide.dev) (`stroke-width: 2`, `fill="none"`).
- **CAUTION (Import Collisions):** Watch for naming collisions between Lucide icon exports and heavyweight React components. For example, always alias `import { Calendar as CalendarIcon } from 'lucide-react'` when working in views that also import the monthly date-picker component `import { Calendar } from '@/components/ui/calendar'`. Mounting a full calendar component in an icon slot blows out card layouts.

### 4.4 Zero-Value State Discipline for Metric Cards
- **PROHIBITED:** Never style a metric card or stat counter with positive green indicators (`text-status-success` or green icon containers) when its value is `0` (e.g. `0 Active Leases`).
- **MANDATORY:** When count is `0`, both the icon badge (`bg-secondary text-muted-foreground border border-border`) and status label (`text-muted-foreground`) must remain neutral. They switch to active green (`bg-[#183925] text-[#98c379] border-[#235836]` / `text-status-success`) strictly when count is greater than zero (`> 0`).

### 4.5 Anti-Full-Bleed Layouts & 50/50 Symmetric Column Splits
- **PROHIBITED:** Never stretch simple forms or small inputs across an unconstrained 1000px+ card. Stretched 600px dropdowns and time pickers in an empty void look amateur and degrade usability.
- **MANDATORY:** When building forms paired with live overviews or rules lists (e.g., Schedule Rule Builder + Configured Policies list), use an equal 50/50 two-column grid on desktop screens (`grid-cols-1 lg:grid-cols-2 gap-6`). This keeps input controls at ergonomic, natural widths (~180px–300px), prevents giant empty voids, and maintains symmetric balance.

### 4.6 Typography Consistency & Anti-Pill Fatigue
- **PROHIBITED:** Never wrap every piece of text in colorful pill badges. Overusing pills produces visual noise and destroys visual hierarchy.
- **MANDATORY:** Standardize badge/pill font sizes strictly to `text-[10px]` or `text-xs` with mono font (`font-mono`). Rely on clean typographic contrast, subtle secondary text (`text-muted-foreground`), and status indicator dots (`h-1.5 w-1.5 rounded-full`) instead of nested pill badges.

### 4.7 Floating Widget Safety & Toast Insets
- **MANDATORY:** Ensure toast viewports (`ToastViewport`) and floating overlays maintain safe viewport insets (e.g., `right-16 sm:right-16`) to prevent overlapping floating trigger buttons or agentation widgets in the bottom-right screen corner.

### 4.8 Data Scalability Pattern (> 10 Items)
- **MANDATORY:** Never assume collections will only contain 1–3 items. Always implement search filtering, scrollable bounded containers (`max-h-60 overflow-y-auto` or `max-h-96`), or pagination for lists that can grow to 10+ items.

---

## 5. Verification & Quality Gate

Before declaring any work complete or opening a PR:

1. **Go Format & Race Tests:**
   ```bash
   gofmt -s -w . && test -z "$(gofmt -l .)"
   go test -race ./...
   ```
2. **Web TypeScript & Bundle Build:**
   ```bash
   bun run --cwd web build
   ```
   Must pass with zero TypeScript errors (`tsc -b`) and successful Vite bundling.
3. **Smoke & Responsive Verification:**
   - Verify CLI commands produce clean, bounded output in standard terminal dimensions (80 and 120 columns).
   - Verify web views remain glitch-free and properly aligned with sidebars open or on narrow laptop viewports.
