package kbviewer

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	keyBindings []key.Binding
}

func New() Model {
	return Model{}
}

func (m *Model) AddKeyBindings(keyBindings ...key.Binding) {
	m.keyBindings = append(m.keyBindings, keyBindings...)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	parts := []string{}

	for _, binding := range m.keyBindings {
		renderedBinding := m.renderKeyBinding(binding)

		if len(parts) == 0 {
			parts = append(parts, renderedBinding)
		} else {
			parts = append(parts, " ", renderedBinding)
		}
	}

	return lipgloss.JoinHorizontal(0, parts...)
}

func (m *Model) renderKeyBinding(binding key.Binding) string {
	keyStyle := lipgloss.NewStyle().Bold(true)
	descriptionStyle := lipgloss.NewStyle()

	key := keyStyle.Render(fmt.Sprintf("[%s]", binding.Help().Key))
	description := descriptionStyle.Render(binding.Help().Desc)

	return lipgloss.JoinHorizontal(0, key, " ", description)
}
