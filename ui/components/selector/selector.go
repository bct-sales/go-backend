package selector

import (
	"log/slog"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Renderable interface {
	Render() string
}

type Model[T Renderable] struct {
	options       []T
	selectedIndex int
	hasFocus      bool
}

func New[T Renderable](options []T) Model[T] {
	return Model[T]{
		options:       options,
		selectedIndex: 0,
		hasFocus:      false,
	}
}

func (m Model[T]) View() string {
	selectedOption := m.options[m.selectedIndex]
	buttonStyle := m.getButtonStyle()

	return lipgloss.JoinHorizontal(
		0,
		buttonStyle.Render("← "),
		m.getOptionStyle().Render(selectedOption.Render()),
		buttonStyle.Render(" →"),
	)
}

func (m Model[T]) getButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().Background(lipgloss.Color("AAAAFF"))
}

func (m Model[T]) getOptionStyle() lipgloss.Style {
	style := lipgloss.NewStyle().Width(m.longestOptionLength() + 2).AlignHorizontal(lipgloss.Center)

	if m.hasFocus {
		style = style.Background(lipgloss.Color("#AAAAFF"))
	}

	return style
}

func (m *Model[T]) longestOptionLength() int {
	result := 0

	for _, option := range m.options {
		renderedOptionLength := len(option.Render())

		if renderedOptionLength > result {
			result = renderedOptionLength
		}
	}

	return result
}

func (m Model[T]) Update(message tea.Msg) (Model[T], tea.Cmd) {
	slog.Debug("Selector received message", slog.Any("message type", reflect.TypeOf(message)), slog.Any("message", message))

	switch message := message.(type) {
	case tea.KeyMsg:
		return m.onKeyPressed(message)

	default:
		return m, nil
	}
}

func (m Model[T]) onKeyPressed(message tea.KeyMsg) (Model[T], tea.Cmd) {
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

func (m Model[T]) selectPreviousOption() (Model[T], tea.Cmd) {
	newSelectedIndex := (m.selectedIndex - 1 + len(m.options)) % len(m.options)
	slog.Debug("Selector: previous option", slog.Int("old selected index", m.selectedIndex), slog.Int("new selected index", newSelectedIndex))
	m.selectedIndex = newSelectedIndex
	return m, nil
}

func (m Model[T]) selectNextOption() (Model[T], tea.Cmd) {
	newSelectedIndex := (m.selectedIndex + 1) % len(m.options)
	slog.Debug("Selector: next option", slog.Int("old selected index", m.selectedIndex), slog.Int("new selected index", newSelectedIndex))
	m.selectedIndex = newSelectedIndex
	return m, nil
}

func (m Model[T]) Init() tea.Cmd {
	return nil
}

func (m *Model[T]) Focus() tea.Cmd {
	m.hasFocus = true
	return nil
}

func (m *Model[T]) Blur() {
	m.hasFocus = false
}
