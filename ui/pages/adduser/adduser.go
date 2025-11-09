package adduser

import (
	"bctbackend/ui/components/selector"
	"bctbackend/ui/pages"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	pages.Page
	components Components
	tabIndex   int
	back       tea.Model
}

type Components struct {
	userID   textinput.Model
	role     selector.Model
	password textinput.Model
}

type Focusable interface {
	Blur()
	Focus() tea.Cmd
}

func New(back tea.Model) Model {
	model := Model{
		components: Components{
			userID:   textinput.New(),
			role:     selector.New([]string{"Admin", "Seller", "Cashier"}),
			password: textinput.New(),
		},
		tabIndex: 0,
		back:     back,
	}

	model.components.userID.Focus()

	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	resultingCommands := []tea.Cmd{}

	{
		m, cmd := model.components.userID.Update(message)
		model.components.userID = m
		resultingCommands = append(resultingCommands, cmd)
	}

	{
		m, cmd := model.components.role.Update(message)
		model.components.role = m
		resultingCommands = append(resultingCommands, cmd)
	}

	{
		m, cmd := model.components.password.Update(message)
		model.components.password = m
		resultingCommands = append(resultingCommands, cmd)
	}

	switch message := message.(type) {
	case tea.KeyMsg:
		updated, c := model.onKeyPressed(message)
		resultingCommands = append(resultingCommands, c)
		return updated, tea.Batch(resultingCommands...)
	}

	return model, tea.Batch(resultingCommands...)
}

// onKeyPressed handles the tea.KeyMsg message
func (m Model) onKeyPressed(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "esc":
		return m.back, m.back.Init()

	case "tab", "down":
		return m.moveFocusToNextComponent()

	case "shift+tab", "up":
		return m.moveFocusToPreviousComponent()

	default:
		return m, nil
	}
}

func (m *Model) getFocusable(index int) Focusable {
	switch index {
	case 0:
		return &m.components.userID
	case 1:
		return &m.components.role
	case 2:
		return &m.components.password
	default:
		panic("invalid index")
	}
}

func (m *Model) getFocusableCount() int {
	return 3
}

func (m Model) moveFocusToPreviousComponent() (Model, tea.Cmd) {
	m.getFocusable(m.tabIndex).Blur()
	m.tabIndex = (m.tabIndex - 1 + m.getFocusableCount()) % m.getFocusableCount()
	cmd := m.getFocusable(m.tabIndex).Focus()

	return m, cmd
}

func (m Model) moveFocusToNextComponent() (Model, tea.Cmd) {
	m.getFocusable(m.tabIndex).Blur()
	m.tabIndex = (m.tabIndex + 1) % m.getFocusableCount()
	cmd := m.getFocusable(m.tabIndex).Focus()

	return m, cmd
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		0,
		"User ID",
		m.components.userID.View(),
		"Role",
		m.components.role.View(),
		"Password",
		m.components.password.View(),
	)
}

func (m Model) StatusBar() string {
	return ""
}

func (m Model) Title() string {
	return m.RenderTitle("Add User")
}
