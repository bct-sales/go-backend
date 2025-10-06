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

type RoleId struct {
	ID
}

func NewRoleId(id ID) RoleId {
	if id != AdminRoleId && id != SellerRoleId && id != CashierRoleId {
		panic(fmt.Sprintf("invalid role id: %d", id))
	}

	return RoleId{ID: id}
}

func NewAdminRoleId() RoleId {
	return NewRoleId(AdminRoleId)
}

func NewSellerRoleId() RoleId {
	return NewRoleId(SellerRoleId)
}

func NewCashierRoleId() RoleId {
	return NewRoleId(CashierRoleId)
}

func ListRoles() []RoleId {
	return []RoleId{
		NewAdminRoleId(),
		NewSellerRoleId(),
		NewCashierRoleId(),
	}
}

func (roleId RoleId) Name() string {
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

func ParseRole(role string) (RoleId, error) {
	switch role {
	case "admin":
		return RoleId{ID: AdminRoleId}, nil
	case "seller":
		return RoleId{ID: SellerRoleId}, nil
	case "cashier":
		return RoleId{ID: CashierRoleId}, nil
	default:
		return RoleId{}, fmt.Errorf("unknown role %s: %w", role, dberr.ErrNoSuchRole)
	}
}

func (roleId RoleId) IsAdmin() bool {
	return roleId.ID == AdminRoleId
}

func (roleId RoleId) IsSeller() bool {
	return roleId.ID == SellerRoleId
}

func (roleId RoleId) IsCashier() bool {
	return roleId.ID == CashierRoleId
}

type RoleVisitor[T any] interface {
	Admin() T
	Seller() T
	Cashier() T
}

func VisitRole[T any](roleId RoleId, visitor RoleVisitor[T]) T {
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

func (roleId RoleId) IsValid() bool {
	return roleId.ID == AdminRoleId || roleId.ID == SellerRoleId || roleId.ID == CashierRoleId
}
