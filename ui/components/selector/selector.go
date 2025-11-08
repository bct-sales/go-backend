package selector

import (
	"log/slog"
	"reflect"

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
	slog.Debug("Viewing selector", slog.Int("selectedIndex", m.selectedIndex))
	selectedOption := m.options[m.selectedIndex]
	buttonStyle := m.getButtonStyle()

	return lipgloss.JoinHorizontal(
		0,
		buttonStyle.Render("← "),
		m.getOptionStyle().Render(selectedOption),
		buttonStyle.Render(" →"),
	)
}

func (m Model) getButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().Background(lipgloss.Color("AAAAFF"))
}

func (m Model) getOptionStyle() lipgloss.Style {
	style := lipgloss.NewStyle().Width(m.longestOptionLength() + 2).AlignHorizontal(lipgloss.Center)

	if m.hasFocus {
		style = style.Background(lipgloss.Color("#AAAAFF"))
	}

	return style
}

func (m *Model) longestOptionLength() int {
	result := 0

	for _, option := range m.options {
		if len(option) > result {
			result = len(option)
		}
	}

	return result
}

func (m Model) Update(message tea.Msg) (Model, tea.Cmd) {
	slog.Debug("Selector received message", slog.Any("message type", reflect.TypeOf(message)), slog.Any("message", message))

	switch message := message.(type) {
	case tea.KeyMsg:
		return m.onKeyPressed(message)

	default:
		return m, nil
	}
}

func (m Model) onKeyPressed(message tea.KeyMsg) (Model, tea.Cmd) {
	if m.hasFocus {
		slog.Debug("Selector processes key message", slog.String("key", message.String()))

		switch message.String() {
		case "left":
			return m.selectPreviousOption()
		case "right":
			return m.selectNextOption()
		default:
			return m, nil
		}
	} else {
		return m, nil
	}
}

func (m Model) selectPreviousOption() (Model, tea.Cmd) {
	newSelectedIndex := (m.selectedIndex - 1 + len(m.options)) % len(m.options)
	slog.Debug("Selector: previous option", slog.Int("old selected index", m.selectedIndex), slog.Int("new selected index", newSelectedIndex))
	m.selectedIndex = newSelectedIndex
	return m, nil
}

func (m Model) selectNextOption() (Model, tea.Cmd) {
	newSelectedIndex := (m.selectedIndex + 1) % len(m.options)
	slog.Debug("Selector: next option", slog.Int("old selected index", m.selectedIndex), slog.Int("new selected index", newSelectedIndex))
	m.selectedIndex = newSelectedIndex
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
