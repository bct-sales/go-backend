package users

import (
	"bctbackend/ui/components/statusbar"
	"bctbackend/ui/pages"
	"bctbackend/ui/pages/adduser"

	tea "github.com/charmbracelet/bubbletea"
)

type DefaultMode struct {
	model     *Model
	statusBar *statusbar.Model
}

func NewDefaultMode(model *Model) pages.Mode {
	statusBar := statusbar.New()
	statusBar.AddKeyBinding("P", "Set password")
	statusBar.AddKeyBinding("+", "Add user")

	return &DefaultMode{
		model:     model,
		statusBar: statusBar,
	}
}

func (mode *DefaultMode) HandleUserInput(message tea.KeyMsg) (pages.PageContents, tea.Cmd) {
	model := mode.model

	switch message.String() {
	case "q":
		return model, tea.Quit

	case "down":
		model.usersView.MoveDown()
		return model, nil

	case "up":
		mode.model.usersView.MoveUp()
		return model, nil

	case "P":
		model.mode = NewSetPasswordMode(model)
		return model, nil

	case "+":
		back := func() tea.Msg {
			return mode.model.RequestSwitchToPage(mode.model)()
		}

		return model, mode.model.RequestSwitchToPage(adduser.New(back))

	default:
		return model, nil
	}
}

func (mode *DefaultMode) View() string {
	return mode.model.usersView.View()
}

func (mode *DefaultMode) StatusBar() string {
	return mode.statusBar.View()
}
