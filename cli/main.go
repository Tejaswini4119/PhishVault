package main

import (
	"fmt"
	"os"
	"strings" // Added strings import

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput" // Added textinput import
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Model ---

type state int

const (
	stateLogin state = iota // New Start State
	stateMenu
	stateRunning
	stateViewingOutput
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
	focusIndex    int // 0=User, 1=Pass, 2=Submit
	loginError    string

	// Menu
	menuItems []menuItem
	cursor    int

	// Common
	spinner      spinner.Model
	viewport     viewport.Model
	output       string
	windowWidth  int
	windowHeight int
	isReady      bool
}

func initialModel() model {
	// 1. Setup Spinner
	s := spinner.New()
	s.Spinner = spinner.Globe // Changed spinner type
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	// 2. Setup Login Inputs
	tiUser := textinput.New()
	tiUser.Placeholder = "Username"
	tiUser.Focus()
	tiUser.CharLimit = 32
	tiUser.Width = 30
	tiUser.TextStyle = lipgloss.NewStyle().Foreground(ColorSecondary)

	tiPass := textinput.New()
	tiPass.Placeholder = "Password"
	tiPass.EchoMode = textinput.EchoPassword
	tiPass.CharLimit = 32
	tiPass.Width = 30
	tiPass.TextStyle = lipgloss.NewStyle().Foreground(ColorSecondary)

	// 3. Setup Menu
	items := []menuItem{
		{"► STATUS DASHBOARD", "View running containers and health", checkDockerStatus},
		{"► START SYSTEM", "Bring up all PhishVault services", startInfrastructure},
		{"► STOP SYSTEM", "Shutdown all services", stopInfrastructure},
		{"► RUN INTELLIGENCE TEST", "Execute full analysis pipeline", runAnalysisTest},
		{"-----------------", "", nil}, // Separator
		{"● OPEN NEO4J", "Open Graph Database (Neo4j Browser)", func() tea.Cmd { return openBrowser("http://localhost:7474") }},
		{"● OPEN MINIO", "Open Object Storage (MinIO Console)", func() tea.Cmd { return openBrowser("http://localhost:9001") }},
		{"● OPEN RABBITMQ", "Open Message Broker (Management UI)", func() tea.Cmd { return openBrowser("http://localhost:15672") }},
		{"-----------------", "", nil},
		{"EXIT", "Quit application", nil},
	}

	return model{
		state:         stateLogin, // Initial state is Login
		usernameInput: tiUser,
		passwordInput: tiPass,
		menuItems:     items,
		spinner:       s,
	}
}

// --- Init ---

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, textinput.Blink) // Added textinput.Blink
}

// --- Update ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// Window Resize
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.isReady {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight-5)
			m.viewport.YPosition = headerHeight + 1
			m.viewport.HighPerformanceRendering = false
			m.isReady = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight - 5
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c": // Moved ctrl+c here
			return m, tea.Quit
		}

		// LOGIN STATE
		if m.state == stateLogin {
			switch msg.String() {
			case "enter":
				if m.focusIndex == 2 {
					// Check Credentials (phishvault / phishvault-tp2)
					if m.usernameInput.Value() == "phishvault" && m.passwordInput.Value() == "phishvault-tp2" {
						m.state = stateMenu
						m.loginError = "" // Clear any previous error
					} else {
						m.loginError = "Access Denied: Invalid Credentials"
						m.passwordInput.SetValue("") // Clear password
					}
				} else if m.focusIndex < 2 {
					m.focusIndex++
				}
			case "tab", "down":
				m.focusIndex = (m.focusIndex + 1) % 3
			case "shift+tab", "up":
				m.focusIndex = (m.focusIndex - 1 + 3) % 3
			}

			// Handle Inputs
			if m.focusIndex == 0 {
				m.usernameInput.Focus()
				m.passwordInput.Blur()
				m.usernameInput, cmd = m.usernameInput.Update(msg)
				cmds = append(cmds, cmd)
			} else if m.focusIndex == 1 {
				m.usernameInput.Blur()
				m.passwordInput.Focus()
				m.passwordInput, cmd = m.passwordInput.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.usernameInput.Blur()
				m.passwordInput.Blur()
			}
			return m, tea.Batch(cmds...)
		}

		// MENU STATE
		if m.state == stateMenu {
			switch msg.String() {
			case "q": // 'q' only quits from menu
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					// Skip separators
					if m.menuItems[m.cursor].title == "-----------------" {
						m.cursor--
					}
				}
			case "down", "j":
				if m.cursor < len(m.menuItems)-1 {
					m.cursor++
					// Skip separators
					if m.menuItems[m.cursor].title == "-----------------" {
						m.cursor++
					}
				}
			case "enter":
				selected := m.menuItems[m.cursor]
				if selected.title == "EXIT" {
					return m, tea.Quit
				}
				if selected.cmd == nil {
					return m, nil // Separator or dummy
				}

				// If specific command needs different state:
				if strings.Contains(selected.title, "OPEN") {
					// Browser commands don't change screen state, just run
					return m, selected.cmd()
				}

				m.state = stateRunning
				m.output = "" // Clear previous output
				return m, tea.Batch(selected.cmd(), m.spinner.Tick)
			}
		}

		// VIEW/RUNNING STATE
		if m.state == stateViewingOutput || m.state == stateRunning {
			switch msg.String() {
			case "esc", "backspace":
				m.state = stateMenu
			default:
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case spinner.TickMsg:
		if m.state == stateRunning { // Only update spinner if in running state
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case CommandOutputMsg:
		m.state = stateViewingOutput
		m.output = string(msg)
		m.viewport.SetContent(m.output)
		m.viewport.GotoBottom()

	case CommandErrorMsg:
		m.state = stateViewingOutput
		m.output = fmt.Sprintf("Error: %v", msg)
		m.viewport.SetContent(m.output)
	}

	return m, tea.Batch(cmds...)
}

// --- View ---

func (m model) View() string {
	if !m.isReady {
		return "\n  Initializing..."
	}

	var s string

	// HEADER / LOGO
	s += m.headerView() + "\n"

	// CONTENT
	if m.state == stateLogin {
		s += m.loginView()
	} else if m.state == stateMenu {
		s += MenuHeaderStyle.Render(" OPERATIONS MENU ") + "\n\n"
		for i, item := range m.menuItems {
			if item.title == "-----------------" {
				s += lipgloss.NewStyle().Foreground(ColorBorder).SetString("  --------------------------------").Render() + "\n"
				continue
			}

			if m.cursor == i {
				s += MenuSelectedStyle.Render(item.title) + " " + lipgloss.NewStyle().Foreground(ColorAccent).Render(item.desc) + "\n"
			} else {
				s += MenuItemStyle.Render(item.title) + "\n"
			}
		}
	} else if m.state == stateRunning {
		s += fmt.Sprintf("\n %s Processing request...\n\n", m.spinner.View())
	} else if m.state == stateViewingOutput {
		s += OutputPanelStyle.Width(m.windowWidth - 4).Render(m.viewport.View())
	}

	s += "\n" + m.footerView()
	return AppStyle.Render(s)
}

func (m model) loginView() string {
	var button string
	if m.focusIndex == 2 {
		button = lipgloss.NewStyle().Foreground(ColorBg).Background(ColorPrimary).Render("[ LOGIN ]")
	} else {
		button = lipgloss.NewStyle().Foreground(ColorFg).Render("[ Login ]")
	}

	errView := ""
	if m.loginError != "" {
		errView = lipgloss.NewStyle().Foreground(ColorError).Render("\n\n" + m.loginError)
	}

	form := fmt.Sprintf(
		"Credentials Required\n\n%s\n\n%s\n\n%s%s",
		m.usernameInput.View(),
		m.passwordInput.View(),
		button,
		errView,
	)

	return LoginBoxStyle.Render(form)
}

func (m model) headerView() string {
	logo := `
██████╗ ██╗  ██╗██╗███████╗██╗  ██╗██╗   ██╗ █████╗ ██╗   ██╗██╗  ████████╗
██╔══██╗██║  ██║██║██╔════╝██║  ██║██║   ██║██╔══██╗██║   ██║██║  ╚══██╔══╝
██████╔╝███████║██║███████╗███████║██║   ██║███████║██║   ██║██║     ██║
██╔═══╝ ██╔══██║██║╚════██║██╔══██║╚██╗ ██╔╝██╔══██║██║   ██║██║     ██║
██║     ██║  ██║██║███████║██║  ██║ ╚████╔╝ ██║  ██║╚██████╔╝███████╗██║
╚═╝     ╚═╝  ╚═╝╚═╝╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝
`
	return LogoStyle.Render(logo)
}

func (m model) footerView() string {
	if m.state == stateLogin {
		return StatusBarStyle.Render("Tab to switch • Enter to submit • Ctrl+C to quit")
	} else if m.state == stateMenu {
		return StatusBarStyle.Render("Use ↑/↓ to navigate • Enter to select • q to quit")
	} else if m.state == stateViewingOutput {
		return StatusBarStyle.Render("Esc to return to menu • q to quit")
	}
	return ""
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Fatal Error: %v", err) // Changed error message
		os.Exit(1)
	}
}
