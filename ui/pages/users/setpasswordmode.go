package users

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"log/slog"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SetPasswordMode struct {
	textInput textinput.Model
}

func NewSetPasswordMode() Mode {
	textInput := textinput.New()
	textInput.Focus()
	textInput.Prompt = "Enter new password> "

	return SetPasswordMode{
		textInput: textInput,
	}
}

func (mode SetPasswordMode) HandleUserInput(model Model, message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "enter":
		index := model.usersView.Selected()
		updatedUser := model.users[index]
		password := mode.textInput.Value()

		model.mode = NewDefaultMode()
		return model, mode.requestUpdateUserPassword(model.Database, updatedUser.UserID, password)

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

func (mode SetPasswordMode) View(model *Model) string {
	return model.usersView.View()
}

func (mode SetPasswordMode) StatusBar(model *Model) string {
	return mode.textInput.View()
}

func (mode *SetPasswordMode) requestUpdateUserPassword(database *sql.DB, userID models.ID, password string) tea.Cmd {
	return func() tea.Msg {
		if err := queries.UpdateUserPassword(database, userID, password); err != nil {
			slog.Error("Failed to update user password", slog.Any("error", err))

			return &databaseErrorMessage{
				err:     err,
				message: "failed to update user password",
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
