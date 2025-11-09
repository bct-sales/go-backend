package pages

import (
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Page struct {
	Database   *sql.DB
	ScreenSize Size
}

func New(database *sql.DB) Page {
	return Page{
		Database:   database,
		ScreenSize: NewSize(0, 0),
	}
}

func (page Page) Init() tea.Cmd {
	return nil
}

func (page Page) Update(message tea.Msg) (Page, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		return page.onWindowResized(message)

	default:
		return page, nil
	}
}

// onWindowResized handles the tea.WindowSizeMsg message
func (page Page) onWindowResized(message tea.WindowSizeMsg) (Page, tea.Cmd) {
	screenWidth := message.Width
	screenHeight := message.Height
	page.ScreenSize = NewSize(screenWidth, screenHeight)

	return page, nil
}

func (page *Page) RenderTitle(title string) string {
	titleStyle := lipgloss.NewStyle().Width(page.ScreenSize.Width).AlignHorizontal(lipgloss.Center).Background(lipgloss.Color("#AAAAFF"))

	return titleStyle.Render(title)
}
