package models

const (
	AdminRoleId   Id = 1
	SellerRoleId  Id = 2
	CashierRoleId Id = 3
)

type Role struct {
	RoleId Id
	Name   string
}

func NewRole(
	id Id,
	name string) *Role {

	return &Role{
		RoleId: id,
		Name:   name,
	}
}
