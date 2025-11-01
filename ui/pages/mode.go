package pages

import tea "github.com/charmbracelet/bubbletea"

type Mode interface {
	HandleUserInput(message tea.KeyMsg) (tea.Model, tea.Cmd)
	View() string
}
