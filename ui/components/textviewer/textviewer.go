package textviewer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	text  string
	style lipgloss.Style
}

func New() Model {
	return Model{
		style: lipgloss.NewStyle(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	return m.style.Render(m.text)
}
