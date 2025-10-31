package ui

import tea "github.com/charmbracelet/bubbletea"

func Start() error {
	program := tea.NewProgram(newRootModel(), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		return err
	}

	return nil
}

type RootModel struct {
	screenWidth  int
	screenHeight int
}

func newRootModel() tea.Model {
	return &RootModel{}
}

func (m *RootModel) Init() tea.Cmd {
	return nil
}

func (m *RootModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		return m.onWindowResized(message)

	case tea.KeyMsg:
		return m.onKeyPressed(message)
	}

	return m, nil
}

// onWindowResized handles the tea.WindowSizeMsg message
func (m *RootModel) onWindowResized(message tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.screenWidth = message.Width
	m.screenHeight = message.Height

	return m, nil
}

// onKeyPressed handles the tea.KeyMsg message
func (m *RootModel) onKeyPressed(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "q":
		return m, tea.Quit
	}
}

func (m *RootModel) View() string {
	return "Hello world"
}
