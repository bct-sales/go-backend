package users

import (
	"bctbackend/ui/components/statusbar"
	"bctbackend/ui/pages/adduser"

	"github.com/charmbracelet/bubbles/key"
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

type keyMap struct {
	quit           key.Binding
	setPassword    key.Binding
	addUser        key.Binding
	selectNext     key.Binding
	selectPrevious key.Binding
}

var defaultKeyMap = keyMap{
	quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	setPassword: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "set password"),
	),
	addUser: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "add user"),
	),
	selectPrevious: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	selectNext: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
}

func (mode DefaultMode) HandleUserInput(model Model, message tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyMap := defaultKeyMap

	switch {
	case key.Matches(message, keyMap.quit):
		return model, tea.Quit

	case key.Matches(message, keyMap.setPassword):
		model.mode = NewSetPasswordMode()
		return model, nil

	case key.Matches(message, keyMap.addUser):
		return mode.onAddUser(model)

	case key.Matches(message, keyMap.selectNext):
		model.usersView.MoveDown()
		return model, nil

	case key.Matches(message, keyMap.selectPrevious):
		model.usersView.MoveUp()
		return model, nil

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
	back := func() (tea.Model, tea.Cmd) {
		return model, model.requestFetchUsers()
	}

	return adduser.New(model.Database, model.ScreenSize, back), nil
}
