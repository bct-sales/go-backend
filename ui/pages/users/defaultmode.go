package users

import (
	"bctbackend/ui/components/kbviewer"
	"bctbackend/ui/pages/adduser"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type DefaultMode struct {
	statusBar kbviewer.Model
}

func NewDefaultMode() Mode {
	keyBindings := DefaultKeyMap

	statusBar := kbviewer.New()
	statusBar.AddKeyBindings(
		keyBindings.addUser,
		keyBindings.removeUser,
		keyBindings.setPassword,
		keyBindings.quit,
	)

	return DefaultMode{
		statusBar: statusBar,
	}
}

type KeyMap struct {
	quit           key.Binding
	setPassword    key.Binding
	addUser        key.Binding
	removeUser     key.Binding
	selectNext     key.Binding
	selectPrevious key.Binding
}

var DefaultKeyMap = KeyMap{
	quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	setPassword: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "Set password"),
	),
	addUser: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "Add user"),
	),
	removeUser: key.NewBinding(
		key.WithKeys("delete"),
		key.WithHelp("del", "Remove user"),
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
	slog.Debug("Key pressed", slog.String("key", message.String()))

	keyMap := DefaultKeyMap

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

	case key.Matches(message, keyMap.removeUser):
		model.mode = NewRemoveUserMode()
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
