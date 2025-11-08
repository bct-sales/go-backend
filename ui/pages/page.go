package pages

import (
	"database/sql"
	"log/slog"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Page struct {
	Database   *sql.DB
	ScreenSize *Size
	Contents   PageContents
}

type PageContents interface {
	Init() tea.Cmd
	Title() string
	Update(message tea.Msg) (PageContents, tea.Cmd)
	View() string
	StatusBar() string
}

func New(database *sql.DB, contents PageContents) *Page {
	return &Page{
		Database:   database,
		ScreenSize: NewSize(0, 0),
		Contents:   contents,
	}
}

func (p *Page) Init() tea.Cmd {
	return p.Contents.Init()
}

func (p *Page) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	slog.Debug("Page.Update", slog.Any("message", message), slog.String("message type", reflect.TypeOf(message).String()))

	switch message := message.(type) {
	case tea.WindowSizeMsg:
		return p.onWindowResized(message)

	case databaseQueryRequestMessage:
		return p.onDatabaseQuery(message.query)

	case switchToPageMessage:
		return p.onSwitchToPage(message.contents)

	default:
		updatedContents, command := p.Contents.Update(message)
		p.Contents = updatedContents
		return p, command
	}
}

func (p *Page) onSwitchToPage(contents PageContents) (tea.Model, tea.Cmd) {
	p.Contents = contents
	command := p.Init()
	return p, command
}

func (p *Page) onDatabaseQuery(query DatabaseQuery) (tea.Model, tea.Cmd) {
	slog.Debug("Database request received; scheduling it", slog.Any("query type", reflect.TypeOf(query)))

	database := p.Database
	command := func() tea.Msg {
		slog.Debug("Performing database query", slog.Any("query type", reflect.TypeOf(query)))
		return query.Perform(database)
	}

	return p, command
}

// onWindowResized handles the tea.WindowSizeMsg message
func (p *Page) onWindowResized(message tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	screenWidth := message.Width
	screenHeight := message.Height
	p.ScreenSize = NewSize(screenWidth, screenHeight)

	updatedContents, command := p.Contents.Update(message)
	p.Contents = updatedContents

	return p, command
}

func (p *Page) View() string {
	title := p.Contents.Title()
	mainView := p.Contents.View()
	statusBar := p.Contents.StatusBar()

	titleHeight := lipgloss.Height(title)
	mainViewHeight := lipgloss.Height(mainView)
	remainingHeight := p.ScreenSize.Height - titleHeight - mainViewHeight
	statusBarStyle := lipgloss.NewStyle().Height(remainingHeight).AlignVertical(lipgloss.Bottom)

	return lipgloss.JoinVertical(0, title, mainView, statusBarStyle.Render(statusBar))
}

func (p *Page) RenderTitle(title string) string {
	titleStyle := lipgloss.NewStyle().Width(p.ScreenSize.Width).AlignHorizontal(lipgloss.Center).Background(lipgloss.Color("#AAAAFF"))

	return titleStyle.Render(title)
}

type PageContentsBase struct{}

func (pc *PageContentsBase) RenderTitle(titleString string) string {
	titleStyle := lipgloss.NewStyle().Background(lipgloss.Color("#AAAAFF")).AlignHorizontal(lipgloss.Center)
	return titleStyle.Render(titleString)
}

func (pc *PageContentsBase) RequestDatabaseQuery(f DatabaseQuery) tea.Cmd {
	slog.Debug("Requesting database query")

	return func() tea.Msg {
		return databaseQueryRequestMessage{f}
	}
}

type DatabaseQuery interface {
	Perform(database *sql.DB) tea.Msg
}

type databaseQueryRequestMessage struct {
	query DatabaseQuery
}

func (pc *PageContentsBase) SwitchToPage(contents PageContents) tea.Cmd {
	return func() tea.Msg {
		return switchToPageMessage{contents}
	}
}

type switchToPageMessage struct {
	contents PageContents
}
