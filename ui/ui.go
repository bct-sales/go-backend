package ui

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/ui/components/usersview"
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
)

func Start(database *sql.DB) error {
	program := tea.NewProgram(newRootModel(database), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		return err
	}

	return nil
}

type RootModel struct {
	database     *sql.DB
	screenWidth  int
	screenHeight int

	usersView *usersview.Model
}

func newRootModel(database *sql.DB) tea.Model {
	return &RootModel{
		database:  database,
		usersView: usersview.New(),
	}
}

func (m *RootModel) Init() tea.Cmd {
	return fetchUsers(m.database)
}

func fetchUsers(database *sql.DB) tea.Cmd {
	return func() tea.Msg {
		var users []*models.User

		if err := queries.GetUsers(database, queries.CollectTo(&users)); err != nil {
			return &DatabaseErrorMessage{
				err:     err,
				message: "failed to fetch users from database",
			}
		}

		return usersFetchedMessage{
			users: users,
		}
	}
}

func (m *RootModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		return m.onWindowResized(message)

	case tea.KeyMsg:
		return m.onKeyPressed(message)

	case usersFetchedMessage:
		m.usersView.SetUsers(message.users)

		return m, nil

	case DatabaseErrorMessage:
		return m, nil
	}

	return m, nil
}

// onWindowResized handles the tea.WindowSizeMsg message
func (m *RootModel) onWindowResized(message tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.screenWidth = message.Width
	m.screenHeight = message.Height

	m.usersView.SetWidth(m.screenWidth)
	m.usersView.SetHeight(m.screenHeight)

	return m, nil
}

// onKeyPressed handles the tea.KeyMsg message
func (m *RootModel) onKeyPressed(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "q":
		return m, tea.Quit

	case "down":
		m.usersView.MoveDown()
		return m, nil

	case "up":
		m.usersView.MoveUp()
		return m, nil
	}

	return m, nil
}

func (m *RootModel) View() string {
	return m.usersView.View()
}

type DatabaseErrorMessage struct {
	message string
	err     error
}

type usersFetchedMessage struct {
	users []*models.User
}
