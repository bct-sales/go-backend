package users

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DefaultMode struct {
	model *Model
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
	return "[P] Change password"
}
