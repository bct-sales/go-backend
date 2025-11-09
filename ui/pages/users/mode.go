package users

import tea "github.com/charmbracelet/bubbletea"

type Mode interface {
	HandleUserInput(model Model, message tea.KeyMsg) (tea.Model, tea.Cmd)
	View(model *Model) string
	StatusBar(model *Model) string
}
