package adduser

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/ui/components/kbviewer"
	"bctbackend/ui/components/selector"
	"bctbackend/ui/components/textviewer"
	"bctbackend/ui/pages"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
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
	role      selector.Model[RoleOption]
	password  textinput.Model
	statusBar tea.Model
}

type Focusable interface {
	Blur()
	Focus() tea.Cmd
}

type KeyBindings struct {
	Cancel   key.Binding
	AddUser  key.Binding
	Next     key.Binding
	Previous key.Binding
}

var DefaultKeyBindings = KeyBindings{
	AddUser: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "add user"),
	),
	Next: key.NewBinding(
		key.WithKeys("tab", "down"),
		key.WithHelp("tab", "down"),
	),
	Previous: key.NewBinding(
		key.WithKeys("shift+tab", "up"),
		key.WithHelp("shift+tab", "up"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
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
	model.components.statusBar = model.createKeyboardBindingViewer()

	return model
}

func (m *Model) createKeyboardBindingViewer() tea.Model {
	keyBindings := DefaultKeyBindings
	viewer := kbviewer.New()

	viewer.AddKeyBindings(
		keyBindings.AddUser,
		keyBindings.Cancel,
	)

	return viewer
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
		return model.showErrorMessage(message.message)
	}

	return model, tea.Batch(resultingCommands...)
}

func (m Model) showErrorMessage(message string) (tea.Model, tea.Cmd) {
	statusBar := textviewer.New()
	statusBar.SetText(message)
	statusBar.SetStyle(lipgloss.NewStyle().Background(lipgloss.Color("#FFAAAA")))
	m.components.statusBar = statusBar
	return m, statusBar.Init()
}

func (m Model) onKeyPressed(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyBindings := DefaultKeyBindings

	switch {
	case key.Matches(message, keyBindings.Cancel):
		return m.back()

	case key.Matches(message, keyBindings.Next):
		return m.moveFocusToNextComponent()

	case key.Matches(message, keyBindings.Previous):
		return m.moveFocusToPreviousComponent()

	case key.Matches(message, keyBindings.AddUser):
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

	statusBar := m.components.statusBar.View()

	return lipgloss.JoinVertical(
		0,
		titleView,
		lipgloss.NewStyle().Height(m.ScreenSize.Height-2).Render(mainView),
		statusBar,
	)
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
		return m.showErrorMessage("Invalid user identifier")
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
