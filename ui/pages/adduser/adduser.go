package adduser

import (
	"bctbackend/ui/pages"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	pages.PageContentsBase
	components Components
	tabIndex   int
	back       tea.Cmd
}

type Components struct {
	userID    textinput.Model
	role      textinput.Model
	password  textinput.Model
	focusable []Focusable
}

type Focusable interface {
	Blur()
	Focus() tea.Cmd
}

func New(back tea.Cmd) *Model {
	model := Model{
		components: Components{
			userID:   textinput.New(),
			role:     textinput.New(),
			password: textinput.New(),
		},
		tabIndex: 0,
		back:     back,
	}

	model.components.focusable = []Focusable{
		&model.components.userID,
		&model.components.role,
		&model.components.password,
	}

	return &model
}

func (m *Model) Init() tea.Cmd {
	return m.components.userID.Focus()
}

func (m *Model) Update(message tea.Msg) (pages.PageContents, tea.Cmd) {
	var resultingModel *Model
	resultingModel = m
	resultingCommands := []tea.Cmd{}

	switch message := message.(type) {
	case tea.KeyMsg:
		m, c := m.onKeyPressed(message)
		resultingModel = m
		resultingCommands = append(resultingCommands, c)
	}

	updatedUserID, userIDCommand := m.components.userID.Update(message)
	updatedRole, roleCommand := m.components.role.Update(message)
	updatedPassword, passwordCommand := m.components.password.Update(message)
	resultingCommands = append(resultingCommands, userIDCommand, roleCommand, passwordCommand)

	m.components.userID = updatedUserID
	m.components.role = updatedRole
	m.components.password = updatedPassword

	return resultingModel, tea.Batch(resultingCommands...)
}

// onKeyPressed handles the tea.KeyMsg message
func (m *Model) onKeyPressed(message tea.KeyMsg) (*Model, tea.Cmd) {
	switch message.String() {
	case "esc":
		return m, m.back

	case "tab", "down":
		cmd := m.moveFocusToNextComponent()
		return m, cmd

	case "shift+tab", "up":
		cmd := m.moveFocusToPreviousComponent()
		return m, cmd

	default:
		return m, nil
	}
}

func (m *Model) moveFocusToPreviousComponent() tea.Cmd {
	m.components.focusable[m.tabIndex].Blur()
	m.tabIndex = (m.tabIndex - 1 + len(m.components.focusable)) % len(m.components.focusable)
	cmd := m.components.focusable[m.tabIndex].Focus()

	return cmd
}

func (m *Model) moveFocusToNextComponent() tea.Cmd {
	m.components.focusable[m.tabIndex].Blur()
	m.tabIndex = (m.tabIndex + 1) % len(m.components.focusable)
	cmd := m.components.focusable[m.tabIndex].Focus()

	return cmd
}

func (m *Model) View() string {
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

func (m *Model) StatusBar() string {
	return ""
}

func (m *Model) Title() string {
	return m.RenderTitle("Add User")
}
