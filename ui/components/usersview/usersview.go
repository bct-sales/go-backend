package usersview

import (
	"bctbackend/database/models"
	"bctbackend/ui/components/listview"
	"fmt"
)

type Model struct {
	view *listview.Model
}

type UserList struct {
	users []*models.User
}

func (list *UserList) Len() int {
	return len(list.users)
}

func (list *UserList) Item(index int) string {
	if index >= list.Len() {
		panic("out of range")
	}

	user := list.users[index]

	return fmt.Sprintf("[%4d] %s", user.UserID, user.RoleID.Name())
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

func (m *Model) SetWidth(width int) {
	m.view.SetWidth(width)
}

func (m *Model) SetHeight(height int) {
	m.view.SetHeight(height)
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
