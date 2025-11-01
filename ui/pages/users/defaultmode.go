package users

import (
	"bctbackend/ui/components/statusbar"
	"bctbackend/ui/pages"
	"bctbackend/ui/pages/adduser"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func (mode *DefaultMode) HandleUserInput(message tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return adduser.New(model.database, model.screenSize), nil

	default:
		return model, nil
	}
}

func (mode *DefaultMode) View() string {
	model := mode.model
	mainView := model.usersView.View()
	statusBar := mode.RenderStatusBar()

	return lipgloss.JoinVertical(0, mainView, statusBar)
}

func (mode *DefaultMode) RenderStatusBar() string {
	return mode.statusBar.View()
}
