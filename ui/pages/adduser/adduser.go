package adduser

import (
	"bctbackend/ui/pages"
	"database/sql"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	pages.Page
	components Components
	tabIndex   int
	back       func() (tea.Model, tea.Cmd)
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

func New(database *sql.DB, screenSize *pages.Size, back func() (tea.Model, tea.Cmd)) tea.Model {
	model := Model{
		Page: pages.Page{
			Database:   database,
			ScreenSize: screenSize,
		},
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

func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var resultingModel tea.Model
	resultingModel = m
	resultingCommands := []tea.Cmd{}

	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m, c := m.onWindowResized(message)
		resultingModel = m
		resultingCommands = append(resultingCommands, c)

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

// onWindowResized handles the tea.WindowSizeMsg message
func (m *Model) onWindowResized(message tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	screenWidth := message.Width
	screenHeight := message.Height

	m.ScreenSize = pages.NewSize(screenWidth, screenHeight)

	return m, nil
}

// onKeyPressed handles the tea.KeyMsg message
func (m *Model) onKeyPressed(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "esc":
		return m.back()

	case "tab":
		cmd := m.moveFocusToNextComponent()
		return m, cmd

	default:
		return m, nil
	}
}

func (m *Model) moveFocusToNextComponent() tea.Cmd {
	m.components.focusable[m.tabIndex].Blur()
	m.tabIndex = (m.tabIndex + 1) % 3
	cmd := m.components.focusable[m.tabIndex].Focus()

	return cmd
}

func (m *Model) View() string {
	title := m.RenderTitle("Add User")

	return lipgloss.JoinVertical(
		0,
		title,
		m.components.userID.View(),
		m.components.role.View(),
		m.components.password.View(),
	)
}
