package views

import "github.com/charmbracelet/lipgloss"

var (
	// -- OpenTUI Gateway Palette --
	// Pure Black & Lime Green
	ColorBg    = lipgloss.Color("#000000")
	ColorFg    = lipgloss.Color("#EEEEEE")
	ColorLime  = lipgloss.Color("#CCFF00") // Electric Lime
	ColorDim   = lipgloss.Color("#444444")
	ColorError = lipgloss.Color("#FF3333")
	ColorInfo  = lipgloss.Color("#00BBFF") // Cyan for info

	// -- Core Layout --
	AppStyle = lipgloss.NewStyle().
			Margin(1, 2).
			Foreground(ColorFg)

	// -- Split Layout Components --

	// Left Side: The Graphic
	LogoContainerStyle = lipgloss.NewStyle().
				Width(50).
				PaddingRight(4).
				Border(lipgloss.NormalBorder(), false, true, false, false). // Right border separator
				BorderForeground(ColorDim).
				Align(lipgloss.Center)

	LogoTextStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			Bold(true)

	// Right Side: The Form
	FormContainerStyle = lipgloss.NewStyle().
				PaddingLeft(4).
				Width(70).
				Align(lipgloss.Left)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			Bold(true).
			MarginBottom(1)

	// Inputs
	InputStyle = lipgloss.NewStyle().
			Foreground(ColorLime).
			Border(lipgloss.NormalBorder(), false, false, true, false). // Bottom border only
			BorderForeground(ColorDim).
			Padding(0, 0) // Tighter padding

	InputPromptStyle = lipgloss.NewStyle().
				Foreground(ColorDim).
				MarginTop(1)

	// Menu / List
	SelectedMenuStyle = lipgloss.NewStyle().
				Foreground(ColorBg).
				Background(ColorLime).
				Bold(true).
				Padding(0, 2).
				MarginLeft(2)

	UnselectedMenuStyle = lipgloss.NewStyle().
				Foreground(ColorFg).
				Padding(0, 2).
				MarginLeft(2)

	// Status
	StatusStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			MarginTop(1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Italic(true)

	// Tree / Results
	NodeHighRisk = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
	NodeSafe     = lipgloss.NewStyle().Foreground(ColorLime)
	NodeNormal   = lipgloss.NewStyle().Foreground(ColorFg)
	NodeInfo     = lipgloss.NewStyle().Foreground(ColorInfo)
)
