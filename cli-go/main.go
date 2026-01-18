package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PhishVault/PhishVault-2/cli-go/integration"
	"github.com/PhishVault/PhishVault-2/cli-go/views"
	"github.com/PhishVault/PhishVault-2/core/domain"
)

// --- State Definitions ---

type state int

const (
	stateLogin state = iota
	stateDashboard
	stateScannerInput
	stateScannerRunning
	stateScannerResults
)

// --- Main Model ---

type model struct {
	state  state
	bridge *integration.Bridge

	// Login Form
	usernameInput textinput.Model
	passwordInput textinput.Model
	focusIndex    int
	loginError    string

	// Dashboard
	menuItems []views.MenuItem
	cursor    int

	// Scanner
	urlInput     textinput.Model
	scanProgress string
	scanResult   domain.SAL
	scanErr      error

	// Common
	spinner      spinner.Model
	viewport     viewport.Model
	output       string
	windowWidth  int
	windowHeight int
	isReady      bool
}

// --- Initialization ---

func initialModel() model {
	// 1. Spinner (Dot)
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(views.ColorLime)

	// 2. Login
	tiUser := textinput.New()
	tiUser.Placeholder = "Enter Operator ID"
	tiUser.Focus()
	tiUser.CharLimit = 32
	tiUser.Width = 30
	tiUser.TextStyle = views.InputStyle

	tiPass := textinput.New()
	tiPass.Placeholder = "Enter Access Token"
	tiPass.EchoMode = textinput.EchoPassword
	tiPass.CharLimit = 32
	tiPass.Width = 30
	tiPass.TextStyle = views.InputStyle

	// 3. Scanner
	tiUrl := textinput.New()
	tiUrl.Placeholder = "https://example.com"
	tiUrl.Width = 40
	tiUrl.TextStyle = views.InputStyle

	items := []views.MenuItem{
		{Title: "SYSTEM STATUS", Desc: "Check container health metrics"},
		{Title: "INTELLIGENCE SCAN", Desc: "Deep analysis of URL artifacts"},
		{Title: "CAMPAIGN DB", Desc: "Search historical campaign graphs"}, // Placeholder
		{Title: "EXIT", Desc: "Terminate session"},
	}

	return model{
		state:         stateLogin,
		usernameInput: tiUser,
		passwordInput: tiPass,
		menuItems:     items,
		spinner:       s,
		urlInput:      tiUrl,
		bridge:        integration.NewBridge(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, textinput.Blink)
}

// --- Update Loop ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.viewport = viewport.New(60, 10)
		m.isReady = true

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if m.bridge != nil {
				m.bridge.Close()
			}
			return m, tea.Quit
		}

		// --- LOGIN ---
		if m.state == stateLogin {
			switch msg.String() {
			case "enter":
				if m.focusIndex == 2 {
					// Verify Creds
					if m.usernameInput.Value() == "phishvault" && m.passwordInput.Value() == "admin" {
						m.state = stateDashboard
					} else {
						m.loginError = "ACCESS DENIED: Invalid Credentials"
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
				if sel.Title == "EXIT" {
					if m.bridge != nil {
						m.bridge.Close()
					}
					return m, tea.Quit
				}
				if sel.Title == "INTELLIGENCE SCAN" {
					m.state = stateScannerInput
					m.urlInput.Focus()
					return m, textinput.Blink
				}
				// Other items unimplemented
				m.output = "Module '" + sel.Title + "' is currently OFFLINE / Placeholder."
				m.viewport.SetContent(m.output)
			}
		}

		// --- SCANNER INPUT ---
		if m.state == stateScannerInput {
			switch msg.String() {
			case "enter":
				if m.urlInput.Value() != "" {
					m.state = stateScannerRunning
					m.scanProgress = "Initiating core engines..."
					return m, tea.Batch(runScan(m.bridge, m.urlInput.Value()), m.spinner.Tick)
				}
			case "esc":
				m.state = stateDashboard
			}
			m.urlInput, cmd = m.urlInput.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		// --- SCANNER RESULTS ---
		if m.state == stateScannerResults {
			if msg.String() == "esc" {
				m.state = stateDashboard
			}
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ScanProgressMsg:
		m.scanProgress = string(msg)
		return m, nil

	case ScanCompleteMsg:
		m.state = stateScannerResults
		if msg.Err != nil {
			m.scanResult = domain.SAL{Verdict: "ERROR", URL: m.urlInput.Value()} // Fallback
			m.scanProgress = "Error: " + msg.Err.Error()
		} else {
			m.scanResult = msg.SAL
		}
	}

	return m, tea.Batch(cmds...)
}

// --- View ---

func (m model) View() string {
	if !m.isReady {
		return "Initializing security protocols..."
	}

	switch m.state {
	case stateLogin:
		return views.RenderLogin(m.usernameInput, m.passwordInput, m.focusIndex, m.loginError)
	case stateDashboard:
		return views.RenderDashboard(m.menuItems, m.cursor, m.output, m.viewport.View())
	case stateScannerInput:
		return views.RenderScannerInput(m.urlInput.View())
	case stateScannerRunning:
		return views.RenderScannerRunning(m.spinner, m.scanProgress)
	case stateScannerResults:
		return views.RenderScannerResults(m.scanResult)
	}

	return "Unknown State"
}

// --- Async Commands ---

type ScanProgressMsg string
type ScanCompleteMsg struct {
	SAL domain.SAL
	Err error
}

func runScan(b *integration.Bridge, url string) tea.Cmd {
	return func() tea.Msg {
		// Simulate steps if we want progress updates (requires callback in Bridge, skipping for now)
		// Directly call bridge
		// We could send a progress msg first if we split this into steps

		time.Sleep(500 * time.Millisecond) // UI effect

		sal, err := b.ScanURL(url)
		return ScanCompleteMsg{SAL: sal, Err: err}
	}
}

// --- Entrypoint ---

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
