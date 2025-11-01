package users

import (
	"bctbackend/database/models"
	"bctbackend/ui/cell"
	"bctbackend/ui/components/usersview"
	"bctbackend/ui/pages"
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	database   *sql.DB
	screenSize *pages.Size
	users      *cell.Observable[[]*models.User]
	usersView  *usersview.Model
	mode       pages.Mode
}

func New(database *sql.DB) tea.Model {
	usersView := usersview.New()
	users := cell.NewObservable[[]*models.User](nil)
	users.AddObserver(func() { usersView.SetUsers(users.Get()) })

	model := Model{
		database:   database,
		screenSize: pages.NewSize(0, 0),
		users:      users,
		usersView:  usersView,
	}

	model.mode = NewDefaultMode(&model)

	return &model
}

func (m *Model) Init() tea.Cmd {
	return fetchUsers(m.database)
}

func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		return m.onWindowResized(message)

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

// onWindowResized handles the tea.WindowSizeMsg message
func (m *Model) onWindowResized(message tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	screenWidth := message.Width
	screenHeight := message.Height

	m.screenSize = pages.NewSize(screenWidth, screenHeight)

	m.usersView.SetWidth(screenWidth)
	m.usersView.SetHeight(screenHeight - 2)

	return m, nil
}

// onKeyPressed handles the tea.KeyMsg message
func (m *Model) onKeyPressed(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.mode.HandleUserInput(message)
}

func (m *Model) View() string {
	titleStyle := lipgloss.NewStyle().Width(m.screenSize.Width).AlignHorizontal(lipgloss.Center).Background(lipgloss.Color("#AAAAFF"))
	title := titleStyle.Render("Users")
	mainView := m.mode.View()

	return lipgloss.JoinVertical(0, title, mainView)
}
