package usersview

import (
	"bctbackend/database/models"
	"bctbackend/ui/components/listview"
	"bctbackend/ui/pages"

	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	view       listview.Model
	users      []*models.User
	screenSize pages.Size
}

type UserList struct {
	users       []*models.User
	screenWidth int
}

func (list *UserList) Len() int {
	return len(list.users)
}

func (list *UserList) RenderItem(index int, selected bool) string {
	if index >= list.Len() {
		panic("out of range")
	}

	userIDStyle := lipgloss.NewStyle().Width(5).AlignHorizontal(lipgloss.Right)
	roleStyle := lipgloss.NewStyle().Width(10).AlignHorizontal(lipgloss.Right)
	passwordStyle := lipgloss.NewStyle().Width(10).AlignHorizontal(lipgloss.Left)

	user := list.users[index]
	userIDView := userIDStyle.Render(user.UserID.String())
	roleView := roleStyle.Render(user.RoleID.Name())
	passwordView := passwordStyle.Render(user.Password)
	lastActivityView := list.renderLastActivity(user)
	userView := lipgloss.JoinHorizontal(0, userIDView, roleView, " ", passwordView, " ", lastActivityView)

	rowStyle := lipgloss.NewStyle().Width(list.screenWidth)
	if selected {
		rowStyle = rowStyle.Background(lipgloss.Color("#AAAAAA"))
	}

	return rowStyle.Render(userView)
}

func (list *UserList) renderLastActivity(user *models.User) string {
	lastActivityStyle := lipgloss.NewStyle()

	var lastActivityString string
	if user.LastActivity != nil {
		lastActivityString = user.LastActivity.FormattedDateTime()
	} else {
		lastActivityString = "N/A"
	}

	return lastActivityStyle.Render(lastActivityString)
}

func New() Model {
	users := UserList{}

	return Model{
		view: listview.New(&users),
	}
}

func (m *Model) View() string {
	return m.view.View()
}

func (m *Model) SetSize(size pages.Size) {
	m.screenSize = size
	m.view.SetHeight(size.Height)
	m.view.SetList(m.createUserListAdapter())
}

func (m *Model) SetUsers(users []*models.User) {
	m.users = users
	wrapper := m.createUserListAdapter()

	m.view.SetList(wrapper)
}

func (m *Model) createUserListAdapter() listview.List {
	return &UserList{users: m.users, screenWidth: m.screenSize.Width}
}

func (m *Model) Selected() int {
	return m.view.Selected()
}

func (m *Model) MoveDown() {
	m.view.MoveDown()
}

func (m *Model) MoveUp() {
	m.view.MoveUp()
}
