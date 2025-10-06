package models

import (
	dberr "bctbackend/database/errors"
	"fmt"
)

const (
	AdminRoleID   ID     = 1
	SellerRoleID  ID     = 2
	CashierRoleID ID     = 3
	AdminName     string = "admin"
	SellerName    string = "seller"
	CashierName   string = "cashier"
)

type RoleID struct {
	ID
}

func NewRoleID(id ID) RoleID {
	if id != AdminRoleID && id != SellerRoleID && id != CashierRoleID {
		panic(fmt.Sprintf("invalid role id: %d", id))
	}

	return RoleID{ID: id}
}

func NewAdminRoleID() RoleID {
	return NewRoleID(AdminRoleID)
}

func NewSellerRoleID() RoleID {
	return NewRoleID(SellerRoleID)
}

func NewCashierRoleID() RoleID {
	return NewRoleID(CashierRoleID)
}

func ListRoles() []RoleID {
	return []RoleID{
		NewAdminRoleID(),
		NewSellerRoleID(),
		NewCashierRoleID(),
	}
}

func (roleID RoleID) Name() string {
	switch roleID.ID {
	case AdminRoleID:
		return AdminName
	case SellerRoleID:
		return SellerName
	case CashierRoleID:
		return CashierName
	default:
		panic(fmt.Sprintf("unknown role id: %d", roleID.ID))
	}
}

func ParseRole(role string) (RoleID, error) {
	switch role {
	case "admin":
		return RoleID{ID: AdminRoleID}, nil
	case "seller":
		return RoleID{ID: SellerRoleID}, nil
	case "cashier":
		return RoleID{ID: CashierRoleID}, nil
	default:
		return RoleID{}, fmt.Errorf("unknown role %s: %w", role, dberr.ErrNoSuchRole)
	}
}

func (roleID RoleID) IsAdmin() bool {
	return roleID.ID == AdminRoleID
}

func (roleID RoleID) IsSeller() bool {
	return roleID.ID == SellerRoleID
}

func (roleID RoleID) IsCashier() bool {
	return roleID.ID == CashierRoleID
}

type RoleVisitor[T any] interface {
	Admin() T
	Seller() T
	Cashier() T
}

func VisitRole[T any](roleID RoleID, visitor RoleVisitor[T]) T {
	switch roleID.ID {
	case AdminRoleID:
		return visitor.Admin()
	case SellerRoleID:
		return visitor.Seller()
	case CashierRoleID:
		return visitor.Cashier()
	default:
		panic(fmt.Sprintf("unknown role id: %d", roleID.ID))
	}
}

func (roleID RoleID) IsValid() bool {
	return roleID.ID == AdminRoleID || roleID.ID == SellerRoleID || roleID.ID == CashierRoleID
}
