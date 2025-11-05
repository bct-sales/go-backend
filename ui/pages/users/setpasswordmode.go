package users

import (
	"bctbackend/ui/pages"

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

func (mode *SetPasswordMode) HandleUserInput(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	model := mode.model

	switch message.String() {
	case "enter":
		index := model.usersView.Selected()
		updatedUser := model.users.Get()[index]
		password := mode.textInput.Value()

		model.mode = NewDefaultMode(model)
		return model, updateUserPassword(model.Database, updatedUser.UserID, password)

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
