package models

import "fmt"

const (
	AdminRoleId   Id     = 1
	SellerRoleId  Id     = 2
	CashierRoleId Id     = 3
	AdminName     string = "admin"
	SellerName    string = "seller"
	CashierName   string = "cashier"
)

type Role struct {
	RoleId Id
	Name   string
}

func NewRole(id Id, name string) *Role {
	return &Role{
		RoleId: id,
		Name:   name,
	}
}

func ParseRole(role string) (Id, error) {
	switch role {
	case "admin":
		return AdminRoleId, nil
	case "seller":
		return SellerRoleId, nil
	case "cashier":
		return CashierRoleId, nil
	default:
		return 0, fmt.Errorf("unknown role: %s", role)
	}
}

type UnknownRoleError struct {
	RoleId Id
}

func (e *UnknownRoleError) Error() string {
	return fmt.Sprintf("unknown role id: %d", e.RoleId)
}

// NameOfRole returns the name of the role with the given id.
// It returns UnknownRoleError if the role id is unknown.
func NameOfRole(roleId Id) (string, error) {
	switch roleId {
	case AdminRoleId:
		return AdminName, nil
	case SellerRoleId:
		return SellerName, nil
	case CashierRoleId:
		return CashierName, nil
	default:
		return "", &UnknownRoleError{RoleId: roleId}
	}
}

func IsValidRole(roleId Id) bool {
	return roleId == AdminRoleId || roleId == SellerRoleId || roleId == CashierRoleId
}
