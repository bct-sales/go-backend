package models

const (
	AdminRoleId   = 1
	SellerRoleId  = 2
	CashierRoleId = 3
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
