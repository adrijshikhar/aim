# AIM Design System

> **Theme**: Atom One Dark  
> **Engine**: [Charm Lipgloss](https://github.com/charmbracelet/lipgloss)  
> **Reference Implementation**: [`internal/tui/styles.go`](file:///Users/nemesis/Projects/my-projects/aim/internal/tui/styles.go)

---

## 1. Design Principles & Vision

AIM (AI Multiplexer) is a high-density, developer-first command-line multiplexer and quota tracker for AI coding agents (`agy`, `gemini`, `codex`, `claude`). Its visual system is built upon four foundational tenets:

1. **Low Visual Fatigue (Atom One Dark Foundation)**  
   Developers spend hours in terminal environments. AIM avoids high-saturation neon colors, stark contrast extremes, or jarring full-screen background washes. It adheres to the classic **Atom One Dark** syntax palette—warm, balanced, and chalky.
2. **Unified TUI & CLI Aesthetic**  
   Whether using the interactive Bubble Tea dashboard (`aim`) or executing headless commands (`aim usage`, `aim doctor`, `aim list`), the typography, color semantics, badges, and progress gauges share the identical Lipgloss token set.
3. **Information Density with Clear Hierarchy**  
   Terminals have constrained real estate. AIM maximizes signal-to-noise ratio by using fixed column widths, subtle secondary text, and bottleneck-first quota reporting.
4. **Declarative & Terminal-Agnostic**  
   All styling is strictly managed through Lipgloss styles. Raw ANSI escapes (`\033[...]`) and manual string-length padding math are prohibited. Truecolor (24-bit RGB) is rendered on supported terminals, automatically degrading cleanly to 256-color or 16-color ANSI, and honoring `NO_COLOR`.

---

## 2. Color Palette & Design Tokens

All tokens are defined in [`internal/tui/styles.go`](file:///Users/nemesis/Projects/my-projects/aim/internal/tui/styles.go).

### 2.1 Text & Neutral Hierarchy

| Token | Hex Code | Visual Sample | Semantic Role | Usage Examples |
|---|---|---|---|---|
| `TextBright` | `#e5e5e5` | ⚪ Bright White | High-contrast foreground, active headers, keycaps | Table headers, modal titles, active selections, `[Key]` caps |
| `TextPrimary` | `#abb2bf` | 🔘 One Dark Gray | Standard readable body text | Normal table rows, body paragraphs, category labels |
| `TextSecondary` | `#9da5b4` | 🔘 UI Neutral | Explanatory text, diagnostics, secondary messages | Diagnostic messages, profile account email details |
| `TextMuted` | `#5c6370` | 🔘 Comment Gray | Faint text, hints, inactive labels | Keybinding descriptions, inactive tabs, timestamps, empty bar slots |
| `TextDim` | `#4b5263` | ⬛ Subtle Charcoal | Borders, brackets, section dividers | Progress bar brackets `[...]`, divider rules `──`, auth methods |

### 2.2 Brand & Accent Colors

| Token | Hex Code | Visual Sample | Semantic Role | Usage Examples |
|---|---|---|---|---|
| `AccentBlue` | `#61afef` | 🔵 Chalky Blue | Signature brand color, active focus, cursor | Cursor indicator `❯`, active tab foreground, drawer borders, modal highlights |
| `AccentCyan` | `#56b6c2` | 🔷 Cyan | Highlighted entities, profile identifiers | Profile names (`default`, `work`), account emails (`user@domain.com`) |
| `AccentPurple` | `#c678dd` | 🟣 Soft Violet | Debugging, spinners, links | CLI loading spinners, `[AIM DEBUG]` log tag, repository URLs |

### 2.3 Surfaces & Backgrounds

| Token | Hex Code | Semantic Role | Usage Examples |
|---|---|---|---|
| `BgTabActive` | `#3e4451` | Active tab pill background, inactive button surface | `TabActiveStyle`, `ModalBtnInactiveStyle` |
| `BgSelectedRow` | `#2c313a` | Cursor row selection highlight | Active item in profile list, selected table rows |

### 2.4 Semantic Status & Quota Colors

AIM uses a 4-tier traffic-light system for quotas and operational health:

| Status Token | Hex Code | Condition / Meaning | Gauge / Badge Example | Usage Context |
|---|---|---|---|---|
| `StatusGreen` | `#98c379` | **OK / Healthy** (> 50% quota remaining, diagnostics passed) | `[OK]` `[85%]` `[████████░░]` | Fresh limits, valid credentials, passed health checks |
| `StatusYellow` | `#e5c07b` | **Warning** (15% – 50% quota, warnings, refreshing) | `[WARN]` `[35%]` `[████░░░░░░]` | Impending limit exhaustion, non-critical warnings |
| `StatusRed` | `#e06c75` | **Critical / Exhausted** (≤ 15% quota, errors, deletion) | `[FAIL]` `[5%]` `[█░░░░░░░░░]` | Exhausted quota, invalid tokens, destructive modals |
| `StatusDim` | `#5c6370` | **Unknown / Offline / Uncredentialed** | `[offline]` `[no credentials]` | Network timeout, missing token, unconfigured agent |

---

## 3. Typography & Hierarchy

Monospace terminal environments do not offer variable font weights or sizing. Hierarchy is achieved through **color value, casing, bolding, faintness, and indentation**.

```
AIM — AI Multiplexer v1.2.0           ← TextBright + StatusGreen + TextSecondary
  [1] Antigravity (agy)  [2] Codex     ← TabActive (AccentBlue on BgTabActive) vs TabInactive (TextMuted)

  PROFILES:                            ← TextBright uppercase section header
  > work          [85%]                ← SelectedRow (TextBright on BgSelectedRow) + GaugeStyle
    personal      [offline]            ← NormalRow (TextPrimary) + StatusDim

  ── Profile Details: work ──          ← TextDim section rule
    Account:      user@corp.com        ← Fixed label (TextSecondary, width: 14) + Value (AccentCyan)
    Claude & GPT: [████████░░] 80%    ← Model category + GaugeBar + Percentage
    Resets At:    2026-09-14 04:00:00  ← Metadata (TextMuted)

  [Enter] Run   [Tab] Switch Agent     ← HintKey (TextBright Bold) + HintLabel (TextMuted)
```

### Hierarchy Breakdown:
1. **Application Banner & Header**:
   - ASCII Art Logo: Styled in `StatusGreen.Bold(true)`.
   - GitHub URL: Styled in `AccentPurple`.
   - Title & Version Tag: `StatusGreen` with version number in `TextSecondary`.
2. **Section Titles & Active Tabs**:
   - Active Tab: `AccentBlue` bold on `BgTabActive` pill background (`Padding(0, 1)`).
   - Inactive Tab: `TextMuted` with bracketed index.
3. **Profile Rows & Selection**:
   - Cursor: `❯` in `AccentBlue.Bold(true)`.
   - Selected Row: `TextBright.Bold(true)` on `BgSelectedRow`.
   - Normal Row: `TextPrimary`.
4. **Metadata & Inspector Rows**:
   - Field Label: `TextSecondary` with fixed `.Width(14)`.
   - Highlighted Values: Profile names and emails in `AccentCyan.Bold(true)`.
   - Secondary Attributes: Auth methods and project IDs in `TextDim`.
5. **Keybinding Bar (Footer)**:
   - Keycap: `[Key]` in `TextBright.Bold(true)`.
   - Action Description: `TextMuted` trailing label with two-space separation.

---

## 4. Reusable UI Components

### 4.1 Status Badges
Status badges standardize diagnostic and quota state across TUI and CLI. They are always formatted with a fixed 8-character width to ensure tabular alignment:

```go
// Render fixed-width badge
badge := tui.GaugeStyleForStatus(st).Width(8).Render(fmt.Sprintf("[%s]", statusStr))
```
- `[OK]    ` — `GaugeGreenStyle` (`#98c379`)
- `[WARN]  ` — `GaugeYellowStyle` (`#e5c07b`)
- `[FAIL]  ` — `GaugeRedStyle` (`#e06c75`)
- `[offline]` — `GaugeDimStyle` (`#5c6370`)

### 4.2 Progress Gauges & Usage Bars
Used to display token and rate-limit quotas visually:

```
[████████░░] 80% (3h 12m)
```
- **Brackets**: `TextDim` (`#4b5263`) `[` and `]`.
- **Filled Segments**: `█` colored dynamically according to threshold:
  - `> 50%`: `StatusGreen` (`#98c379`)
  - `16% – 50%`: `StatusYellow` (`#e5c07b`)
  - `≤ 15%`: `StatusRed` (`#e06c75`)
- **Empty Segments**: `░` colored in `TextMuted` (`#5c6370`).
- **Reset Duration**: Subtle countdown formatted in `TextMuted` (e.g. `(resets in 2h)`).

### 4.3 Dialog Modals
Rendered with rounded borders (`lipgloss.RoundedBorder()`), horizontal centering, and contextual borders:

1. **Delete Confirmation Modal** (`ModalBoxStyle`):
   - Border: `StatusRed` (`#e06c75`).
   - Title: `[!] Confirm Deletion` in `StatusRed.Bold(true)`.
   - Action Buttons:
     - Destructive Action: `ModalBtnActiveStyle` (`TextBright` on `StatusRed` background).
     - Cancel Action: `ModalBtnCancelActiveStyle` (`TextBright` on `AccentBlue` background).
     - Inactive Button: `ModalBtnInactiveStyle` (`TextSecondary` on `BgTabActive`).
2. **Rename Modal** (`RenameModalBoxStyle`):
   - Border: `AccentBlue` (`#61afef`).
   - Title: `✎ Rename Profile` in `AccentBlue.Bold(true)`.
   - TextInput Component: Styled focus cursor with Atom One Dark foreground.
   - Inline Error Banner: `✕ <error>` in `StatusRed.Bold(true)`.

### 4.4 Diagnostics Drawer (`DoctorDrawerStyle`)
- Border: `AccentBlue` rounded border.
- Header: `🩺 Diagnostics: <agent> / <profile>`.
- Content Rows: Fixed-width 8-character badge + 14-character category label + secondary message:
  ```
  [OK]     Binary:        Found codex in PATH
  [OK]     Credentials:   Valid OAuth token present
  [WARN]   Token Age:     Token expires in 12 minutes
  ```

---

## 5. CLI Output Standards

All CLI subcommands (`cmd/aim/`) adhere strictly to the Atom One Dark design system:

### 5.1 Tables (`aim usage`)
- Rendered using `lipgloss/table`.
- **Header Row**: Styled with `tui.TextBright.Bold(true)`.
- **Data Rows**: Styled with `tui.TextPrimary`.
- **Status Cells**: Delegated to `tui.GaugeStyleForStatus(status).Render(str)`.
- **Bars**: Rendered via `renderCLIBar` with dynamic One Dark thresholds.
- **Spinner**: Indeterminate loader styled in `tui.AccentPurple`.

### 5.2 Diagnostic Reports (`aim doctor`)
- Top-level adapter headers in `tui.TextBright.Bold(true)`.
- Profile headers in `tui.AccentCyan.Bold(true)`.
- Results formatted with fixed-width Lipgloss badges (`Width(8)`) and categories (`Width(14)`).

### 5.3 Profile Listing (`aim list`)
- Header: `=== Configured Profiles ===` in `tui.AccentBlue.Bold(true)`.
- Index & Name: Index in `tui.TextMuted`, profile name in `tui.AccentCyan.Bold(true)`.
- Associated Agents: Tagged in `tui.TextBright` (e.g. `[agy, codex]`).
- Status & Quotas: Rendered with corresponding `tui.GaugeStyleForStatus`.
- Filesystem Paths: Faint directory paths in `tui.TextDim`.

### 5.4 Debug Logging (`internal/logger`)
- Formatted as:
  ```
  [AIM DEBUG] 15:04:05.123 Listing profiles (agent="agy")
  ```
  - `[AIM DEBUG]`: Faint violet tag (`#c678dd`, `debugTagStyle`).
  - Timestamp: Faint muted text (`debugTimeStyle`).
  - Log Message: Standard output text.

---

## 6. Implementation Rules for Contributors

When adding new commands, views, or TUI components to AIM, enforce the following engineering rules:

> [!CAUTION]
> **Never use manual string length math on styled strings.**  
> Formats like `fmt.Sprintf("%-14s", styledStr)` or `len(styledStr)` measure the raw byte count—including invisible ANSI escape sequences. This creates severe column misalignment bugs. **Always use Lipgloss width formatting**:
> ```go
> // INCORRECT
> badge := badgeStyle.Render("[OK]")
> line := fmt.Sprintf("%-8s %s", badge, msg) // BROKEN: ANSI bytes corrupt column width
> 
> // CORRECT
> badge := badgeStyle.Width(8).Render("[OK]") // Declarative, accurate terminal width
> line := fmt.Sprintf("%s %s", badge, msg)
> ```

> [!IMPORTANT]
> **No hardcoded ANSI codes or 16-color numbers.**  
> Do not use `\033[...]` or `lipgloss.Color("2")`. Always import and use the semantic tokens from `internal/tui` (`tui.StatusGreen`, `tui.TextPrimary`, etc.).

> [!TIP]
> **Preserve raw text patterns for testability.**  
> Smoke tests (`test/smoke_test.sh`) and CLI assertions grep stdout for tokens like `[OK]`, `[WARN]`, or profile names. Ensure the underlying text content remains uncorrupted when wrapped in Lipgloss styles.
