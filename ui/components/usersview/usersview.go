package usersview

import (
	"bctbackend/database/models"
	"bctbackend/ui/components/listview"
	"bctbackend/ui/pages"

	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	view  listview.Model
	users []*models.User
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
	passwordStyle := lipgloss.NewStyle().Width(20).AlignHorizontal(lipgloss.Left)

	user := list.users[index]
	userIDView := userIDStyle.Render(user.UserID.String())
	roleView := roleStyle.Render(user.RoleID.Name())
	passwordView := passwordStyle.Render(user.Password)
	userView := lipgloss.JoinHorizontal(0, userIDView, roleView, " ", passwordView)

	if selected {
		style := lipgloss.NewStyle().Background(lipgloss.Color("#AAAAAA"))
		userView = style.Render(userView)
	}

	return userView
}

func New() *Model {
	users := UserList{}

	return &Model{
		view: listview.New(&users),
	}
}

func (m *Model) View() string {
	return m.view.View()
}

func (m *Model) SetSize(size pages.Size) {
	m.view.SetHeight(size.Height)
}

func (m *Model) SetUsers(users []*models.User) {
	wrapper := UserList{users: users}

	m.view.SetList(&wrapper)
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
