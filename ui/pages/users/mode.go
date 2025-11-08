package users

import tea "github.com/charmbracelet/bubbletea"

type Mode interface {
	HandleUserInput(model Model, message tea.KeyMsg) (Model, tea.Cmd)
	View(model Model) string
	StatusBar(model Model) string
}
