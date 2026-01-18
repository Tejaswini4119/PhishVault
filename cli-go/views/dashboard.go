package views

import (
	"github.com/charmbracelet/lipgloss"
)

type MenuItem struct {
	Title string
	Desc  string
}

// RenderDashboard displays the main menu and system status
func RenderDashboard(items []MenuItem, cursor int, output string, viewportContent string) string {

	// Left Side: Status / Pyramid
	left := `
          +S
         +SS+
        +SSSS+
       +SSSSSS+
      +SSSSSSSS+
     +SSSSSSSSSS+
    +SSSSSSSSSSSS+
   -SSSSSSSSSSSSSS-
  -SSSSSSSSSSSSSSSS-
 -SSSSSSSSSSSSSSSSSS-
+SSSSSSSSSSSSSSSSSSSS+
+SSSSSSSSSSSSSSSSSSSSSS+
SSSSSSSSSSSSSSSSSSSSSSSS
`
	leftView := LogoContainerStyle.Render(
		LogoTextStyle.Render(left) + "\n\n" +
			lipgloss.NewStyle().Foreground(ColorLime).Render("SYSTEM ONLINE") + "\n" +
			lipgloss.NewStyle().Foreground(ColorDim).Render("Engine: Idle"),
	)

	// Right Side: Menu
	menu := TitleStyle.Render("OPERATIONS CONSOLE") + "\n\n"

	for i, item := range items {
		if cursor == i {
			menu += SelectedMenuStyle.Render("> "+item.Title) +
				lipgloss.NewStyle().Foreground(ColorLime).PaddingLeft(2).Render(item.Desc) + "\n"
		} else {
			menu += UnselectedMenuStyle.Render(item.Title) + "\n"
		}
	}

	// Output Area (if any)
	outputArea := ""
	if output != "" {
		outputArea = "\n" + lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(ColorDim).
			Width(68).
			Render(
				lipgloss.NewStyle().Foreground(ColorDim).Render("--- TERMINAL OUTPUT ---")+"\n"+
					lipgloss.NewStyle().Foreground(ColorInfo).Render(viewportContent),
			)
	}

	rightView := FormContainerStyle.Render(menu + outputArea)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
}
