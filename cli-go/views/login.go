package views

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// RenderLogin returns the string representation of the login screen
func RenderLogin(userField, passField textinput.Model, focusIndex int, err string) string {

	// Create a dynamic login button style based on focus
	btn := "[ Login ]"
	if focusIndex == 2 {
		btn = lipgloss.NewStyle().
			Foreground(ColorBg).
			Background(ColorLime).
			Bold(true).
			Render("[ LOGIN ]")
	} else {
		btn = lipgloss.NewStyle().
			Foreground(ColorDim).
			Render("[ Login ]")
	}

	leftView := LogoContainerStyle.Render(
		LogoTextStyle.Render(`
    ____  __  ____      __      __            ____ 
   / __ \/ / / / /_____/ /_  __|  |  __  ____/ / /_
  / /_/ / /_/ / / ___/ __ \/ / / | / / /_/ / / __/
 / ____/ __  / (__  ) / / /_/ /  |/ / __  / / /_  
/_/   /_/ /_/_/____/_/ /_/\__,_/|__/_/_/_/_/\__/  
                                      v2.0
`) + "\n\n" +
			lipgloss.NewStyle().Foreground(ColorDim).Render("Advanced Intelligence Framework"),
	)

	rightView := FormContainerStyle.Render(
		TitleStyle.Render("SECURE ACCESS GATEWAY") + "\n" +
			lipgloss.NewStyle().Foreground(ColorDim).Render("Identity verification required for access.") + "\n\n" +

			InputPromptStyle.Render("OPERATOR ID") + "\n" +
			userField.View() + "\n" +

			InputPromptStyle.Render("ACCESS TOKEN") + "\n" +
			passField.View() + "\n\n" +

			btn + "\n\n" +
			lipgloss.NewStyle().Foreground(ColorError).Render(err),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
}
