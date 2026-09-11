package tui

import (
	"github.com/aim-cli/aim/internal/usage"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Official Atom One Dark Theme
	TextBright    = lipgloss.Color("#e5e5e5") // One Dark bright white / titles / keycaps
	TextPrimary   = lipgloss.Color("#abb2bf") // One Dark default foreground
	TextSecondary = lipgloss.Color("#9da5b4") // One Dark UI neutral
	TextMuted     = lipgloss.Color("#5c6370") // One Dark comment / hint gray
	TextDim       = lipgloss.Color("#4b5263") // One Dark subtle border / divider

	AccentBlue    = lipgloss.Color("#61afef") // One Dark signature chalky blue
	AccentCyan    = lipgloss.Color("#56b6c2") // One Dark cyan
	AccentPurple  = lipgloss.Color("#c678dd") // One Dark purple / violet
	BgTabActive   = lipgloss.Color("#3e4451") // One Dark pill / active tab background
	BgSelectedRow = lipgloss.Color("#2c313a") // One Dark row selection highlight

	// Status Quota Colors (Atom One Dark syntax palette)
	StatusGreen  = lipgloss.Color("#98c379") // One Dark Green - OK
	StatusYellow = lipgloss.Color("#e5c07b") // One Dark Gold / Yellow - Warning
	StatusRed    = lipgloss.Color("#e06c75") // One Dark Coral Red - Critical / Exhausted
	StatusDim    = lipgloss.Color("#5c6370") // One Dark Comment Gray - Unknown / Offline

	// Backward compatibility aliases
	AccentSky = AccentBlue
	Violet    = AccentBlue
	Gray      = TextMuted
	Muted     = TextMuted
	Green     = StatusGreen
	BorderCol = TextDim

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(TextBright).
			Padding(0, 1)

	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentBlue).
			Background(BgTabActive).
			Padding(0, 1)

	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(TextMuted).
				Padding(0, 1)

	SelectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(TextBright).
				Background(BgSelectedRow)

	NormalRowStyle = lipgloss.NewStyle().
			Foreground(TextPrimary)

	CursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentBlue)

	HintKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(TextBright)

	HintLabelStyle = lipgloss.NewStyle().
			Foreground(TextMuted)

	GaugeGreenStyle  = lipgloss.NewStyle().Bold(true).Foreground(StatusGreen)
	GaugeYellowStyle = lipgloss.NewStyle().Bold(true).Foreground(StatusYellow)
	GaugeRedStyle    = lipgloss.NewStyle().Bold(true).Foreground(StatusRed)
	GaugeDimStyle    = lipgloss.NewStyle().Foreground(StatusDim)

	ModalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(StatusRed).
			Padding(1, 2).
			MarginLeft(2)

	ModalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(StatusRed)

	ModalBtnActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(TextBright).
				Background(StatusRed).
				Padding(0, 1)

	ModalBtnInactiveStyle = lipgloss.NewStyle().
				Foreground(TextSecondary).
				Background(BgTabActive).
				Padding(0, 1)

	ModalBtnCancelActiveStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(TextBright).
					Background(AccentBlue).
					Padding(0, 1)

	DoctorDrawerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(AccentBlue).
				Padding(1, 2).
				MarginLeft(2)

	RenameModalBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(AccentBlue).
				Padding(1, 2).
				MarginLeft(2)

	RenameModalTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(AccentBlue)
)

func GaugeStyleForStatus(st usage.Status) lipgloss.Style {
	switch st {
	case usage.StatusOK:
		return GaugeGreenStyle
	case usage.StatusWarning:
		return GaugeYellowStyle
	case usage.StatusCritical, usage.StatusExhausted:
		return GaugeRedStyle
	default:
		return GaugeDimStyle
	}
}
