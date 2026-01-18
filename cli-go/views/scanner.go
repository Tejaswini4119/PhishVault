package views

import (
	"fmt"
	"strings"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// RenderScannerInput shows the URL input field
func RenderScannerInput(urlInputView string) string {
	leftView := LogoContainerStyle.Render(
		TitleStyle.Render("ARCHITECTURAL SCAN") + "\n" +
			lipgloss.NewStyle().Foreground(ColorDim).Render("Initializing Deep Scan Protocol..."),
	)

	rightView := FormContainerStyle.Render(
		TitleStyle.Render("TARGET ACQUISITION") + "\n" +
			lipgloss.NewStyle().Foreground(ColorDim).Render("Enter full URL for analysis.") + "\n\n" +

			InputPromptStyle.Render("TARGET URL") + "\n" +
			urlInputView + "\n\n" +
			lipgloss.NewStyle().Foreground(ColorDim).Render("[Enter] Execute Scan  [Esc] Abort"),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
}

// RenderScannerRunning shows the spinner and progress
func RenderScannerRunning(sp spinner.Model, progress string) string {
	return lipgloss.Place(
		80, 20,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			sp.View(),
			"\n",
			lipgloss.NewStyle().Foreground(ColorLime).Render(progress),
		),
	)
}

// RenderScannerResults displays the SAL content
func RenderScannerResults(sal domain.SAL) string {

	// Verdict Color
	verdictColor := NodeNormal
	if sal.Verdict == "MALICIOUS" {
		verdictColor = NodeHighRisk
	} else if sal.Verdict == "SAFE" {
		verdictColor = NodeSafe
	}

	// Build Tree
	var sb strings.Builder
	sb.WriteString(NodeNormal.Render("◉ " + sal.URL + "\n"))
	sb.WriteString("│\n")

	// Signals
	sb.WriteString("├─ SIGNALS DETECTED:\n")
	if len(sal.Signals) == 0 {
		sb.WriteString("│  └─ " + NodeSafe.Render("No anomalous signals found.") + "\n")
	} else {
		for _, sig := range sal.Signals {
			score := fmt.Sprintf("%.2f", sig.Confidence)
			style := NodeNormal
			if sig.Confidence > 0.8 {
				style = NodeHighRisk
			}
			sb.WriteString(fmt.Sprintf("│  ├─ [%s] %s (Conf: %s)\n",
				style.Render(sig.EngineName),
				sig.SignalKey,
				score,
			))
		}
	}
	sb.WriteString("│\n")

	// Final Verdict
	sb.WriteString("└─ VERDICT: " + verdictColor.Bold(true).Render(sal.Verdict))
	sb.WriteString(fmt.Sprintf("\n   RISK SCORE: %.2f", sal.RiskScore))

	if sal.CampaignID != "" {
		sb.WriteString("\n   CAMPAIGN ID: " + NodeInfo.Render(sal.CampaignID))
	}

	leftView := LogoContainerStyle.Render(
		TitleStyle.Render("SCAN COMPLETE") + "\n\n" +
			verdictColor.Render(sal.Verdict),
	)

	rightView := FormContainerStyle.Render(
		sb.String() + "\n\n" +
			lipgloss.NewStyle().Foreground(ColorDim).Render("[Esc] Return to Console"),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
}
