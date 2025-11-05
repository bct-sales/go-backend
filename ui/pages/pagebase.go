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

func (p *PageBase) AddStatusBar(mainView string, statusBar string) string {
	mainViewHeight := lipgloss.Height(mainView)
	remainingHeight := p.ScreenSize.Height - mainViewHeight
	statusBarStyle := lipgloss.NewStyle().Height(remainingHeight).AlignVertical(lipgloss.Bottom)
	return lipgloss.JoinVertical(0, mainView, statusBarStyle.Render(statusBar))
}
