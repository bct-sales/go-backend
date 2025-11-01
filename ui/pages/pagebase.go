package pages

import (
	"database/sql"

	"github.com/charmbracelet/lipgloss"
)

type PageBase struct {
	Database   *sql.DB
	ScreenSize *Size
}

func (p *PageBase) RenderTitle(title string) string {
	titleStyle := lipgloss.NewStyle().Width(p.ScreenSize.Width).AlignHorizontal(lipgloss.Center).Background(lipgloss.Color("#AAAAFF"))

	return titleStyle.Render(title)
}
