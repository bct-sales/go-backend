package users

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"log/slog"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type RemoveUserMode struct {
	textInput textinput.Model
}

func NewRemoveUserMode() RemoveUserMode {
	textInput := textinput.New()
	textInput.Focus()
	textInput.Prompt = "Confirm user removal by typing yes> "

	return RemoveUserMode{
		textInput: textInput,
	}
}

func (mode RemoveUserMode) HandleUserInput(model Model, message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "enter":
		selectedIndex := model.usersView.Selected()
		selectedUser := model.users[selectedIndex]
		confirmation := mode.textInput.Value()

		if confirmation == "yes" {
			model.mode = NewDefaultMode()
			return model, mode.requestRemoveUser(model.Database, selectedUser.UserID)
		}

		return model, nil

	case "esc":
		model.mode = NewDefaultMode()
		return model, nil

	default:
		updatedTextInput, command := mode.textInput.Update(message)
		mode.textInput = updatedTextInput
		model.mode = mode
		return model, command
	}
}

func (mode RemoveUserMode) View(model *Model) string {
	return model.usersView.View()
}

func (mode RemoveUserMode) StatusBar(model *Model) string {
	return mode.textInput.View()
}

func (mode *RemoveUserMode) requestRemoveUser(database *sql.DB, userID models.ID) tea.Cmd {
	return func() tea.Msg {
		if err := queries.RemoveUserWithID(database, userID); err != nil {
			slog.Error("Failed to remove user", slog.Any("error", err))

			return &databaseErrorMessage{
				err:     err,
				message: "failed to remove user",
			}
		}

		var users []*models.User
		if err := queries.GetUsers(database, queries.CollectTo(&users)); err != nil {
			slog.Error("Failed to fetch updated user information", slog.Any("error", err))

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
