package pages

import tea "github.com/charmbracelet/bubbletea"

type Mode interface {
	HandleUserInput(message tea.KeyMsg) (PageContents, tea.Cmd)
	View() string
	StatusBar() string
}
