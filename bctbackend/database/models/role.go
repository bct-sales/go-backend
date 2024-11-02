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

func NameOfRole(roleId Id) (string, error) {
	switch roleId {
	case AdminRoleId:
		return AdminName, nil
	case SellerRoleId:
		return SellerName, nil
	case CashierRoleId:
		return CashierName, nil
	default:
		return "", fmt.Errorf("unknown role id: %d", roleId)
	}
}
