package main

import "github.com/charmbracelet/lipgloss"

var (
	// -- Colors --
	// Industrial Palette: Dark Slate, Neon Green, Muted Steel
	ColorBg        = lipgloss.Color("#1a1b26") // Deep Night
	ColorFg        = lipgloss.Color("#a9b1d6") // Muted Text
	ColorPrimary   = lipgloss.Color("#00e1b6") // PhishVault Neon Green (Core Brand)
	ColorSecondary = lipgloss.Color("#7aa2f7") // Cyber Blue
	ColorAccent    = lipgloss.Color("#e0af68") // Warning Gold
	ColorBorder    = lipgloss.Color("#414868") // Steel Border
	ColorError     = lipgloss.Color("#f7768e") // Alert Red

	// -- Core Layout --
	AppStyle = lipgloss.NewStyle().
			Margin(1, 2).
			BorderForeground(ColorBorder)

	// -- Components --

	LogoStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			MarginBottom(1).
			Align(lipgloss.Center)

	// Login Box
	LoginBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(ColorSecondary).
			Padding(1, 4).
			Align(lipgloss.Center)

	// Menu
	MenuHeaderStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorBorder).
			MarginBottom(1)

	MenuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(ColorFg)

	MenuSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(ColorPrimary).
				Bold(true).
				SetString("▒▓ ") // Industrial cursor

	// Output Output Panel
	OutputPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder).
				Padding(0, 1).
				MarginTop(1)

	// Footer table or status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorBorder).
			MarginTop(2)
)
