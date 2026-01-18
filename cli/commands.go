package main

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// Messages
type CommandOutputMsg string
type CommandErrorMsg error

// Helper to run a shell command and return output as a Msg
func executeCommand(command string, args ...string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(command, args...)
		// Set dir if needed, but for now root is fine if running from root
		// cmd.Dir = "./"

		out, err := cmd.CombinedOutput()
		if err != nil {
			return CommandErrorMsg(err)
		}
		return CommandOutputMsg(string(out))
	}
}

// Specific Commands

func checkDockerStatus() tea.Cmd {
	return executeCommand("docker", "ps", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
}

func startInfrastructure() tea.Cmd {
	// docker-compose up -d inside deploy/
	return func() tea.Msg {
		cmd := exec.Command("docker", "compose", "up", "-d")
		cmd.Dir = "deploy"
		out, err := cmd.CombinedOutput()
		if err != nil {
			return CommandErrorMsg(err)
		}
		return CommandOutputMsg("Infrastructure Started:\n" + string(out))
	}
}

func stopInfrastructure() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("docker", "compose", "down")
		cmd.Dir = "deploy"
		out, err := cmd.CombinedOutput()
		if err != nil {
			return CommandErrorMsg(err)
		}
		return CommandOutputMsg("Infrastructure Stopped:\n" + string(out))
	}
}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		// Windows specific start
		cmd = exec.Command("cmd", "/c", "start", url)
		if err := cmd.Start(); err != nil {
			return CommandErrorMsg(err)
		}
		// Return a status message that doesn't block UI
		return nil
	}
}

func runAnalysisTest() tea.Cmd {
	// go test ...
	return executeCommand("go", "test", "-v", "-run", "TestOrchestratorIntegration", "./services/analysis/...")
}
