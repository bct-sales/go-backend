package statusbar

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	keyBindings []keyBinding
}

type keyBinding struct {
	Key         string
	Description string
}

func New() *Model {
	return &Model{}
}

func (m *Model) AddKeyBinding(key string, description string) {
	binding := keyBinding{key, description}
	m.keyBindings = append(m.keyBindings, binding)
}

func (m *Model) View() string {
	parts := []string{}

	for _, binding := range m.keyBindings {
		renderedBinding := m.renderKeyBinding(&binding)
		parts = append(parts, renderedBinding)
	}

	return lipgloss.JoinHorizontal(0, parts...)
}

func (m *Model) renderKeyBinding(binding *keyBinding) string {
	keyStyle := lipgloss.NewStyle().Bold(true)
	descriptionStyle := lipgloss.NewStyle()

	key := keyStyle.Render(fmt.Sprintf("[%s]", binding.Key))
	description := descriptionStyle.Render(binding.Description)

	return lipgloss.JoinHorizontal(0, key, " ", description)
}
