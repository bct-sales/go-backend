package pages

import tea "github.com/charmbracelet/bubbletea"

type Mode interface {
	HandleUserInput(message tea.KeyMsg) (tea.Model, tea.Cmd)
	View() string
}

type Size struct {
	Width  int
	Height int
}

func NewSize(width int, height int) *Size {
	return &Size{width, height}
}
