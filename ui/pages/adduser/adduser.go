package adduser

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/ui/components/selector"
	"bctbackend/ui/pages"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	pages.Page
	components   Components
	tabIndex     int
	back         func() (tea.Model, tea.Cmd)
	errorMessage string
}

type Components struct {
	userID   textinput.Model
	role     selector.Model[RoleOption]
	password textinput.Model
}

type Focusable interface {
	Blur()
	Focus() tea.Cmd
}

func New(database *sql.DB, screenSize pages.Size, back func() (tea.Model, tea.Cmd)) Model {
	roles := algorithms.Map(models.ListRoles(), func(roleID models.RoleID) RoleOption { return RoleOption{roleID} })

	model := Model{
		Page: pages.Page{
			Database:   database,
			ScreenSize: screenSize,
		},
		components: Components{
			userID:   textinput.New(),
			role:     selector.New(roles),
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

	case successMessage:
		return model.back()

	case failureMessage:
		model.errorMessage = message.message
		return model, nil
	}

	return model, tea.Batch(resultingCommands...)
}

func (m Model) onKeyPressed(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "esc":
		return m.back()

	case "tab", "down":
		return m.moveFocusToNextComponent()

	case "shift+tab", "up":
		return m.moveFocusToPreviousComponent()

	case "enter":
		return m.onAddUser()

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
	titleView := m.RenderTitle("Add User")

	mainView := lipgloss.JoinVertical(
		0,
		"User ID",
		m.components.userID.View(),
		"Role",
		m.components.role.View(),
		"Password",
		m.components.password.View(),
	)

	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(
		0,
		titleView,
		lipgloss.NewStyle().Height(m.ScreenSize.Height-2).Render(mainView),
		statusBar,
	)
}

func (m *Model) renderStatusBar() string {
	if len(m.errorMessage) == 0 {
		return ""
	}

	style := lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")).Width(m.ScreenSize.Width)
	return style.Render(m.errorMessage)
}

func (m Model) StatusBar() string {
	return ""
}

func (m Model) Title() string {
	return m.RenderTitle("Add User")
}

func (m Model) onAddUser() (tea.Model, tea.Cmd) {
	userID, err := models.ParseID(m.components.userID.Value())
	if err != nil {
		slog.Error("Invalid user identifier while adding user")
		m.errorMessage = "Invalid user id"
		return m, nil
	}

	roleID := m.components.role.GetSelected().roleId
	password := m.components.password.Value()
	createdAt := models.Now()

	command := func() tea.Msg {
		slog.Debug(
			"Adding user",
			slog.Int64("userID", userID.Int64()),
			slog.String("role", roleID.Name()),
			slog.String("password", password),
			slog.String("createdAt", createdAt.FormattedDateTime()),
		)

		err := queries.AddUserWithID(m.Database, userID, roleID, createdAt, nil, password)

		if err != nil {
			slog.Error("Database error while adding new user", slog.Any("error", err))

			if errors.Is(err, dberr.ErrIDAlreadyInUse) {
				return failureMessage{"ID already in use"}
			}

			return failureMessage{"A database error occurred"}
		}

		return successMessage{}
	}

	return m, command
}

type successMessage struct{}

type failureMessage struct {
	message string
}
