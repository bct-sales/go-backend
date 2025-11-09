package users

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/ui/cell"
	"bctbackend/ui/components/usersview"
	"bctbackend/ui/pages"
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	pages.Page
	users     *cell.Observable[[]*models.User]
	usersView *usersview.Model
	mode      Mode
}

func New(database *sql.DB) Model {
	usersView := usersview.New()
	users := cell.NewObservable[[]*models.User](nil)
	users.AddObserver(func() { usersView.SetUsers(users.Get()) })

	model := Model{
		Page:      pages.New(database),
		users:     users,
		usersView: usersView,
		mode:      NewDefaultMode(),
	}

	return model
}

func (model Model) Init() tea.Cmd {
	return model.requestFetchUsers()
}

func (model Model) requestFetchUsers() tea.Cmd {
	return func() tea.Msg {
		var users []*models.User

		if err := queries.GetUsers(model.Database, queries.CollectTo(&users)); err != nil {
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

func (model Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		model.onWindowResized(message)

	case tea.KeyMsg:
		return model.onKeyPressed(message)

	case usersFetchedMessage:
		model.users.Set(message.users)

		return model, nil

	case databaseErrorMessage:
		return model, tea.Quit
	}

	return model, nil
}

func (model Model) onWindowResized(message tea.WindowSizeMsg) (Model, tea.Cmd) {
	screenWidth := message.Width
	screenHeight := message.Height

	model.usersView.SetSize(pages.Size{
		Width:  screenWidth,
		Height: screenHeight - 2,
	})

	return model, nil
}

// onKeyPressed handles the tea.KeyMsg message
func (model Model) onKeyPressed(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	return model.mode.HandleUserInput(model, message)
}

func (model Model) View() string {
	titleView := model.RenderTitle("Users")
	mainView := model.mode.View(&model)
	statusBarView := model.mode.StatusBar(&model)

	titleViewHeight := lipgloss.Height(titleView)
	statusBarHeight := lipgloss.Height(statusBarView)
	remainingHeight := model.ScreenSize.Height - titleViewHeight - statusBarHeight
	mainViewStyle := lipgloss.NewStyle().Height(remainingHeight)

	return lipgloss.JoinVertical(0, titleView, mainViewStyle.Render(mainView), statusBarView)
}
