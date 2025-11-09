package listview

import (
	"github.com/charmbracelet/lipgloss"
)

type List interface {
	Len() int
	RenderItem(index int) string
}

type Model struct {
	list            List
	width           int
	height          int
	firstShown      int
	selected        int
	selectedStyle   lipgloss.Style
	unselectedStyle lipgloss.Style
}

func New(list List) *Model {
	return &Model{
		list:            list,
		height:          0,
		firstShown:      0,
		selected:        0,
		selectedStyle:   lipgloss.NewStyle(),
		unselectedStyle: lipgloss.NewStyle(),
	}
}

func (m *Model) View() string {
	if m.list == nil || m.list.Len() == 0 {
		return ""
	}

	if m.firstShown > m.selected || m.selected >= m.list.Len() {
		panic("bug detected")
	}

	index := m.firstShown
	accumulatedHeight := 0
	accumulatedViews := []string{}

	reachedSelected := false

	for !reachedSelected || (accumulatedHeight < m.height && index < m.list.Len()) {
		styledCurrentItem := m.renderItem(index)

		if index == m.selected {
			reachedSelected = true
		}

		accumulatedViews = append(accumulatedViews, styledCurrentItem)

		currentItemHeight := lipgloss.Height(styledCurrentItem)
		accumulatedHeight += currentItemHeight

		for accumulatedHeight > m.height && len(accumulatedViews) > 1 {
			m.firstShown += 1
			accumulatedHeight -= lipgloss.Height(accumulatedViews[0])
			accumulatedViews = accumulatedViews[1:]
		}

		index += 1
	}

	return lipgloss.NewStyle().Height(m.height).Render(lipgloss.JoinVertical(0, accumulatedViews...))
}

func (m *Model) renderItem(index int) string {
	item := m.list.RenderItem(index)
	isItemSelected := index == m.selected

	var style *lipgloss.Style
	if isItemSelected {
		style = &m.selectedStyle
	} else {
		style = &m.unselectedStyle
	}

	return style.Render(item)
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.layoutUpdated()
}

func (m *Model) SetHeight(height int) {
	m.height = height
	m.layoutUpdated()
}

func (m *Model) SetList(list List) {
	m.list = list
}

func (m *Model) GetList() List {
	return m.list
}

func (m *Model) layoutUpdated() {
	basicStyle := lipgloss.NewStyle().Width(m.width)

	m.unselectedStyle = basicStyle
	m.selectedStyle = basicStyle.Background(lipgloss.Color("#AAAAAA"))
}

func (m *Model) Selected() int {
	return m.selected
}

func (m *Model) MoveDown() {
	if m.selected < m.list.Len()-1 {
		m.selected += 1
	}
}

func (m *Model) MoveUp() {
	if m.selected > 0 {
		m.selected -= 1

		if m.selected < m.firstShown {
			m.firstShown = m.selected
		}
	}
}
