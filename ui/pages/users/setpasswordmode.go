package users

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/ui/pages"
	"database/sql"
	"log/slog"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SetPasswordMode struct {
	model     *Model
	textInput textinput.Model
}

func NewSetPasswordMode(model *Model) pages.Mode {
	textInput := textinput.New()
	textInput.Focus()
	textInput.Prompt = "Enter new password> "

	return &SetPasswordMode{
		model:     model,
		textInput: textInput,
	}
}

func (mode *SetPasswordMode) HandleUserInput(message tea.KeyMsg) (pages.PageContents, tea.Cmd) {
	model := mode.model

	switch message.String() {
	case "enter":
		index := model.usersView.Selected()
		updatedUser := model.users.Get()[index]
		password := mode.textInput.Value()

		model.mode = NewDefaultMode(model)
		return model, mode.requestUpdateUserPassword(updatedUser.UserID, password)

	case "esc":
		model.mode = NewDefaultMode(model)
		return model, nil

	default:
		updatedTextInput, command := mode.textInput.Update(message)
		mode.textInput = updatedTextInput
		return model, command
	}
}

func (mode *SetPasswordMode) View() string {
	return mode.model.usersView.View()
}

func (mode *SetPasswordMode) StatusBar() string {
	return mode.textInput.View()
}

func (mode *SetPasswordMode) requestUpdateUserPassword(userID models.ID, password string) tea.Cmd {
	slog.Debug("Creating request to update user password", slog.Int64("userID", userID.Int64()), slog.String("password", password))
	query := updatePasswordQuery{userID, password}

	return mode.model.RequestDatabaseQuery(&query)
}

type updatePasswordQuery struct {
	userID   models.ID
	password string
}

func (q *updatePasswordQuery) Perform(database *sql.DB) tea.Msg {
	slog.Debug("Updating user password", slog.Int64("userID", q.userID.Int64()), slog.String("password", q.password))
	if err := queries.UpdateUserPassword(database, q.userID, q.password); err != nil {
		slog.Error("Failed to update user password", slog.Any("error", err))

		return &databaseErrorMessage{
			err:     err,
			message: "failed to update user password",
		}
	}

	slog.Debug("Fetching updated user information")
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
