package users

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/ui/components/usersview"
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	database     *sql.DB
	screenWidth  int
	screenHeight int
	usersView    *usersview.Model
}

func New(database *sql.DB) tea.Model {
	return &Model{
		database:  database,
		usersView: usersview.New(),
	}
}

func (m *Model) Init() tea.Cmd {
	return fetchUsers(m.database)
}

func fetchUsers(database *sql.DB) tea.Cmd {
	return func() tea.Msg {
		var users []*models.User

		if err := queries.GetUsers(database, queries.CollectTo(&users)); err != nil {
			return &databaseErrorMessage{
				err:     err,
				message: "failed to fetch users from database",
			}
		}

		return usersFetchedMessage{
			users: users,
		}
	}
}

func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		return m.onWindowResized(message)

	case tea.KeyMsg:
		return m.onKeyPressed(message)

	case usersFetchedMessage:
		m.usersView.SetUsers(message.users)

		return m, nil

	case databaseErrorMessage:
		return m, nil
	}

	return m, nil
}

// onWindowResized handles the tea.WindowSizeMsg message
func (m *Model) onWindowResized(message tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.screenWidth = message.Width
	m.screenHeight = message.Height

	m.usersView.SetWidth(m.screenWidth)
	m.usersView.SetHeight(m.screenHeight)

	return m, nil
}

// onKeyPressed handles the tea.KeyMsg message
func (m *Model) onKeyPressed(message tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m *Model) View() string {
	return m.usersView.View()
}

type databaseErrorMessage struct {
	message string
	err     error
}

type usersFetchedMessage struct {
	users []*models.User
}
