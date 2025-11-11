package ui

import "github.com/charmbracelet/lipgloss"

// Color definitions
var (
	ColorRed       = lipgloss.Color("1")
	ColorGreen     = lipgloss.Color("2")
	ColorYellow    = lipgloss.Color("3")
	ColorBlue      = lipgloss.Color("4")
	ColorMagenta   = lipgloss.Color("5")
	ColorCyan      = lipgloss.Color("6")
	ColorGray      = lipgloss.Color("8")
	ColorWhite     = lipgloss.Color("15")
	ColorDefault   = lipgloss.Color("7")
)

// Style definitions
var (
	// Header style
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorCyan).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Padding(0, 1)

	// Title style
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorCyan).
		Padding(0, 1)

	// List styles
	ListStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBlue).
		Padding(1)

	ListItemNormalStyle = lipgloss.NewStyle().
		Foreground(ColorDefault)

	ListItemSelectedStyle = lipgloss.NewStyle().
		Background(ColorGray).
		Foreground(ColorWhite)

	// Preview styles
	PreviewStyle = lipgloss.NewStyle().
		Padding(1)

	PreviewTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorCyan)

	// Status bar style
	StatusBarStyle = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Background(ColorGray).
		Padding(0, 1)

	// File status styles
	StagedStyle = lipgloss.NewStyle().
		Foreground(ColorGreen).
		Bold(true)

	UnstagedStyle = lipgloss.NewStyle().
		Foreground(ColorRed).
		Bold(true)

	UntrackedStyle = lipgloss.NewStyle().
		Foreground(ColorYellow).
		Bold(true)

	// Message styles
	SuccessStyle = lipgloss.NewStyle().
		Foreground(ColorGreen).
		Bold(true)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ColorRed).
		Bold(true)

	WarningStyle = lipgloss.NewStyle().
		Foreground(ColorYellow).
		Bold(true)

	InfoStyle = lipgloss.NewStyle().
		Foreground(ColorBlue)

	// Help style
	HelpStyle = lipgloss.NewStyle().
		Faint(true).
		Foreground(ColorGray)

	// Checkbox styles
	CheckedStyle = lipgloss.NewStyle().
		Foreground(ColorGreen).
		Bold(true)

	UncheckedStyle = lipgloss.NewStyle().
		Foreground(ColorGray)
)

// FileStatusStyle returns the appropriate style for a file status
func FileStatusStyle(statusSymbol string) lipgloss.Style {
	switch statusSymbol {
	case "+":
		return StagedStyle
	case "-":
		return UnstagedStyle
	case "?":
		return UntrackedStyle
	default:
		return lipgloss.NewStyle()
	}
}

// FileStatusColor returns the appropriate color for a file status
func FileStatusColor(statusSymbol string) lipgloss.Color {
	switch statusSymbol {
	case "+":
		return ColorGreen
	case "-":
		return ColorRed
	case "?":
		return ColorYellow
	default:
		return ColorDefault
	}
}
