package models

import (
	dberr "bctbackend/database/errors"
	"fmt"
)

const (
	AdminRoleId   ID     = 1
	SellerRoleId  ID     = 2
	CashierRoleId ID     = 3
	AdminName     string = "admin"
	SellerName    string = "seller"
	CashierName   string = "cashier"
)

type RoleID struct {
	ID
}

func NewRoleId(id ID) RoleID {
	if id != AdminRoleId && id != SellerRoleId && id != CashierRoleId {
		panic(fmt.Sprintf("invalid role id: %d", id))
	}

	return RoleID{ID: id}
}

func NewAdminRoleID() RoleID {
	return NewRoleId(AdminRoleId)
}

func NewSellerRoleID() RoleID {
	return NewRoleId(SellerRoleId)
}

func NewCashierRoleID() RoleID {
	return NewRoleId(CashierRoleId)
}

func ListRoles() []RoleID {
	return []RoleID{
		NewAdminRoleID(),
		NewSellerRoleID(),
		NewCashierRoleID(),
	}
}

func (roleId RoleID) Name() string {
	switch roleId.ID {
	case AdminRoleId:
		return AdminName
	case SellerRoleId:
		return SellerName
	case CashierRoleId:
		return CashierName
	default:
		panic(fmt.Sprintf("unknown role id: %d", roleId.ID))
	}
}

func ParseRole(role string) (RoleID, error) {
	switch role {
	case "admin":
		return RoleID{ID: AdminRoleId}, nil
	case "seller":
		return RoleID{ID: SellerRoleId}, nil
	case "cashier":
		return RoleID{ID: CashierRoleId}, nil
	default:
		return RoleID{}, fmt.Errorf("unknown role %s: %w", role, dberr.ErrNoSuchRole)
	}
}

func (roleId RoleID) IsAdmin() bool {
	return roleId.ID == AdminRoleId
}

func (roleId RoleID) IsSeller() bool {
	return roleId.ID == SellerRoleId
}

func (roleId RoleID) IsCashier() bool {
	return roleId.ID == CashierRoleId
}

type RoleVisitor[T any] interface {
	Admin() T
	Seller() T
	Cashier() T
}

func VisitRole[T any](roleId RoleID, visitor RoleVisitor[T]) T {
	switch roleId.ID {
	case AdminRoleId:
		return visitor.Admin()
	case SellerRoleId:
		return visitor.Seller()
	case CashierRoleId:
		return visitor.Cashier()
	default:
		panic(fmt.Sprintf("unknown role id: %d", roleId.ID))
	}
}

func (roleId RoleID) IsValid() bool {
	return roleId.ID == AdminRoleId || roleId.ID == SellerRoleId || roleId.ID == CashierRoleId
}
