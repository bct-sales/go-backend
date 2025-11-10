package users

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/ui/components/usersview"
	"bctbackend/ui/pages"
	"database/sql"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	pages.Page
	users     []*models.User
	usersView usersview.Model
	mode      Mode
}

func New(database *sql.DB) Model {
	usersView := usersview.New()

	model := Model{
		Page:      pages.New(database),
		users:     nil,
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
		return model.onWindowResized(message), nil

	case tea.KeyMsg:
		return model.onKeyPressed(message)

	case usersFetchedMessage:
		model.users = message.users
		model.usersView.SetUsers(model.users)

		return model, nil

	case databaseErrorMessage:
		return model, tea.Quit
	}

	return model, nil
}

func (model Model) onWindowResized(message tea.WindowSizeMsg) tea.Model {
	slog.Debug("Window resized", "message", message)

	screenWidth := message.Width
	screenHeight := message.Height

	model.ScreenSize = pages.Size{Width: screenWidth, Height: screenHeight}

	model.usersView.SetSize(pages.Size{
		Width:  screenWidth,
		Height: screenHeight - 2,
	})

	return model
}

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
