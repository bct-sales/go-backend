package usersview

import (
	"bctbackend/database/models"
	"bctbackend/ui/components/listview"
	"bctbackend/ui/pages"

	"github.com/charmbracelet/lipgloss"
)

const (
	userIDColumnWidth       = 5
	roleColumnWidth         = 10
	lastActivityColumnWidth = 30
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

	user := list.users[index]
	userIDView := list.renderUserID(user.UserID)
	roleView := list.renderRole(user.RoleID)
	passwordView := list.renderPassword(user.Password)
	lastActivityView := list.renderLastActivity(user.LastActivity)
	userView := lipgloss.JoinHorizontal(0, userIDView, roleView, " ", lastActivityView, " ", passwordView)

	rowStyle := lipgloss.NewStyle().Width(list.screenWidth)
	if selected {
		rowStyle = rowStyle.Background(lipgloss.Color("#AAAAAA"))
	}

	return rowStyle.Render(userView)
}

func (list *UserList) renderUserID(userID models.ID) string {
	style := lipgloss.NewStyle().Width(userIDColumnWidth).AlignHorizontal(lipgloss.Right)
	return style.Render(userID.String())
}

func (list *UserList) renderRole(role models.RoleID) string {
	style := lipgloss.NewStyle().Width(roleColumnWidth).AlignHorizontal(lipgloss.Right)
	return style.Render(role.Name())
}

func (list *UserList) renderPassword(password string) string {
	style := lipgloss.NewStyle()
	return style.Render(password)
}

func (list *UserList) renderLastActivity(lastActivity *models.Timestamp) string {
	style := lipgloss.NewStyle().Width(lastActivityColumnWidth)

	var lastActivityString string
	if lastActivity != nil {
		lastActivityString = lastActivity.FormattedDateTime()
	} else {
		lastActivityString = "N/A"
	}

	return style.Render(lastActivityString)
}

func New() Model {
	users := UserList{}

	return Model{
		view: listview.New(&users),
	}
}

func (m Model) View() string {
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
