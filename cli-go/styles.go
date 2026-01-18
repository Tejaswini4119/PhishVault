package main

import "github.com/charmbracelet/lipgloss"

var (
	// -- OpenTUI Gateway Palette --
	// Pure Black & Lime Green
	ColorBg    = lipgloss.Color("#000000")
	ColorFg    = lipgloss.Color("#EEEEEE")
	ColorLime  = lipgloss.Color("#CCFF00") // Electric Lime
	ColorDim   = lipgloss.Color("#444444") // Placeholder
	ColorError = lipgloss.Color("#FF3333")

	// -- Core Layout --
	AppStyle = lipgloss.NewStyle().
			Margin(2, 4).
			Foreground(ColorFg)

	// -- Split Layout Components --

	// Left Side: The Graphic
	LogoContainerStyle = lipgloss.NewStyle().
				Width(60).
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
				Width(60).
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
			Padding(1, 0)

	InputPromptStyle = lipgloss.NewStyle().
				Foreground(ColorDim).
				MarginTop(1)

	// Menu / List
	SelectedMenuStyle = lipgloss.NewStyle().
				Foreground(ColorBg).
				Background(ColorLime).
				Bold(true).
				Padding(0, 1)

	UnselectedMenuStyle = lipgloss.NewStyle().
				Foreground(ColorFg).
				Padding(0, 1)

	// Status
	StatusStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			MarginTop(2)

	// Tree
	NodeHighRisk = lipgloss.NewStyle().Foreground(ColorError)
	NodeSafe     = lipgloss.NewStyle().Foreground(ColorLime)
	NodeNormal   = lipgloss.NewStyle().Foreground(ColorFg)
)
