package selector

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	options       []string
	selectedIndex int
	hasFocus      bool
}

func New(options []string) Model {
	return Model{
		options:       options,
		selectedIndex: 0,
		hasFocus:      false,
	}
}

func (m Model) View() string {
	selectedOption := m.options[m.selectedIndex]

	return lipgloss.JoinHorizontal(
		0,
		"← ",
		selectedOption,
		" →",
	)
}

func (m Model) Update(message tea.Msg) (Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.KeyMsg:
		return m.onKeyPressed(message)

	default:
		return m, nil
	}
}

func (m Model) onKeyPressed(message tea.KeyMsg) (Model, tea.Cmd) {
	switch message.String() {
	case "left":
		return m.selectPreviousOption()
	case "right":
		return m.selectNextOption()
	default:
		return m, nil
	}
}

func (m Model) selectPreviousOption() (Model, tea.Cmd) {
	m.selectedIndex = (m.selectedIndex - 1 + len(m.options)) % len(m.options)
	return m, nil
}

func (m Model) selectNextOption() (Model, tea.Cmd) {
	m.selectedIndex = (m.selectedIndex + 1) % len(m.options)
	return m, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Focus() tea.Cmd {
	m.hasFocus = true
	return nil
}

func (m *Model) Blur() {
	m.hasFocus = false
}
