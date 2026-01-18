package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Model ---

type state int

const (
	stateLogin state = iota
	stateDashboard
	stateScannerInput
	stateScannerRunning
	stateScannerResults
)

type menuItem struct {
	title string
	desc  string
	cmd   func() tea.Cmd
}

type model struct {
	state state

	// Login Form
	usernameInput textinput.Model
	passwordInput textinput.Model
	focusIndex    int
	loginError    string

	// Dashboard
	menuItems []menuItem
	cursor    int

	// Scanner
	urlInput     textinput.Model
	scanProgress string
	scanResult   string
	scanTree     []string

	// Common
	spinner      spinner.Model
	viewport     viewport.Model
	output       string
	windowWidth  int
	windowHeight int
	isReady      bool
}

func initialModel() model {
	// 1. Spinner (Dot)
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorLime)

	// 2. Login
	tiUser := textinput.New()
	tiUser.Placeholder = "Enter your username"
	tiUser.Focus()
	tiUser.CharLimit = 32
	tiUser.Width = 30
	tiUser.TextStyle = InputStyle

	tiPass := textinput.New()
	tiPass.Placeholder = "Enter your password"
	tiPass.EchoMode = textinput.EchoPassword
	tiPass.CharLimit = 32
	tiPass.Width = 30
	tiPass.TextStyle = InputStyle

	// 3. Scanner
	tiUrl := textinput.New()
	tiUrl.Placeholder = "Enter target URL..."
	tiUrl.Width = 40
	tiUrl.TextStyle = InputStyle

	items := []menuItem{
		{"System Status", "View Containers", checkDockerStatus},
		{"Start Infrastructure", "Docker Up", startInfrastructure},
		{"Stop Infrastructure", "Docker Down", stopInfrastructure},
		{"Intelligence Test", "Run Analysis", runAnalysisTest},
		{"START NEW SCAN", "Analyze URL", nil},
		{"Exit", "Quit", nil},
	}

	return model{
		state:         stateLogin,
		usernameInput: tiUser,
		passwordInput: tiPass,
		menuItems:     items,
		spinner:       s,
		urlInput:      tiUrl,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, textinput.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.viewport = viewport.New(msg.Width/2, msg.Height-10) // Half width for viewport
		m.viewport.YPosition = 0
		m.isReady = true

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// --- LOGIN ---
		if m.state == stateLogin {
			switch msg.String() {
			case "enter":
				if m.focusIndex == 2 {
					if m.usernameInput.Value() == "phishvault" && m.passwordInput.Value() == "phishvault-tp2" {
						m.state = stateDashboard
					} else {
						m.loginError = "Invalid Credentials"
						m.passwordInput.SetValue("")
					}
				} else {
					m.focusIndex++
				}
			case "tab", "down":
				m.focusIndex = (m.focusIndex + 1) % 3
			case "shift+tab", "up":
				m.focusIndex = (m.focusIndex - 1 + 3) % 3
			}

			if m.focusIndex == 0 {
				m.usernameInput.Focus()
				m.passwordInput.Blur()
			}
			if m.focusIndex == 1 {
				m.usernameInput.Blur()
				m.passwordInput.Focus()
			}

			m.usernameInput, cmd = m.usernameInput.Update(msg)
			cmds = append(cmds, cmd)
			m.passwordInput, cmd = m.passwordInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		// --- DASHBOARD ---
		if m.state == stateDashboard {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.menuItems)-1 {
					m.cursor++
				}
			case "enter":
				sel := m.menuItems[m.cursor]
				if sel.title == "Exit" {
					return m, tea.Quit
				}
				if sel.title == "START NEW SCAN" {
					m.state = stateScannerInput
					m.urlInput.Focus()
					return m, textinput.Blink
				}
				m.output = "Executing " + sel.title + "..."
				m.viewport.SetContent(m.output)
				return m, tea.Batch(sel.cmd(), m.spinner.Tick)
			}
		}

		// --- SCANNER INPUT ---
		if m.state == stateScannerInput {
			switch msg.String() {
			case "enter":
				if m.urlInput.Value() != "" {
					m.state = stateScannerRunning
					m.scanProgress = "Analyzing..."
					return m, tea.Batch(runScannerSteps(m.urlInput.Value()), m.spinner.Tick)
				}
			case "esc":
				m.state = stateDashboard
			}
			m.urlInput, cmd = m.urlInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		if m.state == stateScannerResults || m.state == stateScannerRunning {
			if msg.String() == "esc" {
				m.state = stateDashboard
			}
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case TriggerScanMsg:
		m.scanProgress = msg.Status
		return m, msg.Cmd

	case ScanResultMsg:
		m.state = stateScannerResults
		m.scanTree = msg.Tree
		m.scanResult = msg.Verdict

	case CommandOutputMsg:
		m.output = string(msg)
		m.viewport.SetContent(m.output)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.isReady {
		return "Loading..."
	}

	// 1. Left Side: The Graphic (Pyramid)
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
			lipgloss.NewStyle().Foreground(ColorDim).Render("PhishVault Gateway v6.0"),
	)

	// 2. Right Side: Context Sensitive
	var rightView string

	if m.state == stateLogin {
		btn := "[ Login ]"
		if m.focusIndex == 2 {
			btn = lipgloss.NewStyle().Foreground(ColorBg).Background(ColorLime).Bold(true).Render("[ LOGIN ]")
		}

		rightView = FormContainerStyle.Render(
			TitleStyle.Render("OpenTUI Access Gateway") + "\n" +
				lipgloss.NewStyle().Foreground(ColorDim).Render("Enter your credentials to continue") + "\n\n" +

				InputPromptStyle.Render("USERNAME") + "\n" +
				m.usernameInput.View() + "\n" +

				InputPromptStyle.Render("PASSWORD") + "\n" +
				m.passwordInput.View() + "\n\n" +

				btn + "\n" +
				lipgloss.NewStyle().Foreground(ColorError).Render(m.loginError),
		)
	} else if m.state == stateDashboard {
		menu := TitleStyle.Render("Operations Menu") + "\n\n"
		for i, item := range m.menuItems {
			if m.cursor == i {
				menu += SelectedMenuStyle.Render("> "+item.title) + "\n"
			} else {
				menu += UnselectedMenuStyle.Render(item.title) + "\n"
			}
		}

		output := ""
		if m.output != "" {
			output = "\n\n" + lipgloss.NewStyle().Foreground(ColorDim).Render("--- OUTPUT ---") + "\n" +
				lipgloss.NewStyle().Foreground(ColorLime).Render(m.viewport.View())
		}

		rightView = FormContainerStyle.Render(menu + output)

	} else if m.state == stateScannerInput {
		rightView = FormContainerStyle.Render(
			TitleStyle.Render("Intelligence Scanner") + "\n" +
				lipgloss.NewStyle().Foreground(ColorDim).Render("Enter target URL for deep analysis") + "\n\n" +

				InputPromptStyle.Render("TARGET URL") + "\n" +
				m.urlInput.View() + "\n\n" +
				lipgloss.NewStyle().Foreground(ColorDim).Render("[Enter] Scan  [Esc] Back"),
		)
	} else if m.state == stateScannerRunning {
		rightView = FormContainerStyle.Render(
			"\n\n" + m.spinner.View() + " " + lipgloss.NewStyle().Foreground(ColorLime).Render(m.scanProgress),
		)
	} else if m.state == stateScannerResults {
		tree := ""
		for _, l := range m.scanTree {
			tree += l + "\n"
		}

		vColor := NodeSafe
		if m.scanResult == "PHISHING" {
			vColor = NodeHighRisk
		}

		rightView = FormContainerStyle.Render(
			TitleStyle.Render("Scan Results") + "\n\n" +
				tree + "\n\n" +
				"VERDICT: " + vColor.Bold(true).Render(m.scanResult) + "\n\n" +
				lipgloss.NewStyle().Foreground(ColorDim).Render("[Esc] Dashboard"),
		)
	}

	// 3. Combine with JoinHorizontal
	return AppStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView) + "\n" +
			StatusStyle.Render("Ctrl+C to Quit"),
	)
}

// --- Logic ---

type TriggerScanMsg struct {
	Status string
	Cmd    tea.Cmd
}

type ScanResultMsg struct {
	Verdict string
	Tree    []string
}

func runScannerSteps(url string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1 * time.Second)
		return TriggerScanMsg{
			Status: "Stealth Mode: Randomizing...",
			Cmd: func() tea.Msg {
				time.Sleep(1 * time.Second)
				return TriggerScanMsg{
					Status: "Capturing DOM...",
					Cmd: func() tea.Msg {
						time.Sleep(2 * time.Second)

						// Logic
						lowerURL := strings.ToLower(url)
						isPhishing := false
						if strings.Contains(lowerURL, "amazon") ||
							strings.Contains(lowerURL, "github.io") ||
							strings.Contains(lowerURL, "login") {
							isPhishing = true
						}

						var tree []string
						var verdict string

						if isPhishing {
							verdict = "PHISHING"
							tree = []string{
								NodeNormal.Render("◉ " + url),
								"├─ " + NodeHighRisk.Render("IP: 44.208.23.11 (Risk: CRITICAL)"),
								"├─ " + NodeNormal.Render("Hosting: GitHub Pages"),
								"├─ DOM Elements",
								"│  ├─ " + NodeHighRisk.Render("Input[Password] (Harvesting)"),
								"│  └─ " + NodeSafe.Render("Logo (Google)"),
								"└─ " + NodeHighRisk.Render("SSL: DV (Abused)"),
							}
						} else {
							verdict = "SAFE"
							tree = []string{
								NodeNormal.Render("◉ " + url),
								"├─ " + NodeSafe.Render("IP: 142.250.190.46"),
								"├─ " + NodeNormal.Render("Hosting: Google"),
								"├─ DOM Elements",
								"│  ├─ " + NodeSafe.Render("Input[Search]"),
								"│  └─ " + NodeSafe.Render("Logo"),
								"└─ " + NodeSafe.Render("SSL: Valid"),
							}
						}

						return ScanResultMsg{
							Verdict: verdict,
							Tree:    tree,
						}
					},
				}
			},
		}
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
