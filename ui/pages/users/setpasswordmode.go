package users

import (
	"bctbackend/ui/pages"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		return model, updateUserPassword(model.database, updatedUser.UserID, password)

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
	model := mode.model
	mainView := model.usersView.View()
	statusBar := mode.RenderStatusBar()

	return lipgloss.JoinVertical(0, mainView, statusBar)
}

func (mode *SetPasswordMode) RenderStatusBar() string {
	return mode.textInput.View()
}
