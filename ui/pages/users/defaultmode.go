package users

import (
	"bctbackend/ui/components/statusbar"
	"bctbackend/ui/pages/adduser"

	tea "github.com/charmbracelet/bubbletea"
)

type DefaultMode struct {
	statusBar *statusbar.Model
}

func NewDefaultMode() Mode {
	statusBar := statusbar.New()
	statusBar.AddKeyBinding("P", "Set password")
	statusBar.AddKeyBinding("+", "Add user")

	return DefaultMode{
		statusBar: statusBar,
	}
}

func (mode DefaultMode) HandleUserInput(model Model, message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "q":
		return model, tea.Quit

	case "down":
		model.usersView.MoveDown()
		return model, nil

	case "up":
		model.usersView.MoveUp()
		return model, nil

	case "P":
		model.mode = NewSetPasswordMode()
		return model, nil

	case "+":
		return mode.onAddUser(model)

	default:
		return model, nil
	}
}

func (mode DefaultMode) View(model *Model) string {
	return model.usersView.View()
}

func (mode DefaultMode) StatusBar(model *Model) string {
	return mode.statusBar.View()
}

func (mode *DefaultMode) onAddUser(model Model) (tea.Model, tea.Cmd) {
	return adduser.New(model), nil
}
