package users

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/ui/cell"
	"bctbackend/ui/components/usersview"
	"bctbackend/ui/pages"
	"database/sql"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	pages.PageContentsBase
	users     *cell.Observable[[]*models.User]
	usersView *usersview.Model
	mode      Mode
}

func New() Model {
	usersView := usersview.New()
	users := cell.NewObservable[[]*models.User](nil)
	users.AddObserver(func() { usersView.SetUsers(users.Get()) })

	model := Model{
		users:     users,
		usersView: usersView,
		mode:      NewDefaultMode(),
	}

	return model
}

func (m Model) Init() tea.Cmd {
	slog.Debug("Initializing users page")
	return m.requestFetchUsers()
}

func (m Model) requestFetchUsers() tea.Cmd {
	slog.Debug("Requested a user fetch")

	return m.RequestDatabaseQuery(&fetchUsersQuery{})
}

type fetchUsersQuery struct{}

func (q *fetchUsersQuery) Perform(database *sql.DB) tea.Msg {
	slog.Debug("Performing user fetch")

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

func (m Model) Update(message tea.Msg) (pages.PageContents, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.onWindowResized(message)

	case tea.KeyMsg:
		return m.onKeyPressed(message)

	case usersFetchedMessage:
		m.users.Set(message.users)

		return m, nil

	case databaseErrorMessage:
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) onWindowResized(message tea.WindowSizeMsg) (Model, tea.Cmd) {
	screenWidth := message.Width
	screenHeight := message.Height

	m.usersView.SetSize(pages.Size{
		Width:  screenWidth,
		Height: screenHeight - 2,
	})

	return m, nil
}

// onKeyPressed handles the tea.KeyMsg message
func (m Model) onKeyPressed(message tea.KeyMsg) (Model, tea.Cmd) {
	return m.mode.HandleUserInput(m, message)
}

func (m Model) View() string {
	return m.mode.View(m)
}

func (m Model) StatusBar() string {
	return m.mode.StatusBar(m)
}

func (m Model) Title() string {
	return m.RenderTitle("Users")
}
