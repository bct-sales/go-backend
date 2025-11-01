package adduser

import (
	"bctbackend/ui/pages"
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	pages.PageBase
}

func New(database *sql.DB, screenSize *pages.Size) tea.Model {
	model := Model{
		PageBase: pages.PageBase{
			Database:   database,
			ScreenSize: screenSize,
		},
	}

	return &model
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		return m.onWindowResized(message)

	case tea.KeyMsg:
		return m.onKeyPressed(message)
	}

	return m, nil
}

// onWindowResized handles the tea.WindowSizeMsg message
func (m *Model) onWindowResized(message tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	screenWidth := message.Width
	screenHeight := message.Height

	m.ScreenSize = pages.NewSize(screenWidth, screenHeight)

	return m, nil
}

// onKeyPressed handles the tea.KeyMsg message
func (m *Model) onKeyPressed(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "q":
		return m, tea.Quit

	default:
		return m, nil
	}
}

func (m *Model) View() string {
	return m.RenderTitle("Add User")
}
