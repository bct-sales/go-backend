package models

import (
	"errors"
	"fmt"
)

const (
	AdminRoleId   Id     = 1
	SellerRoleId  Id     = 2
	CashierRoleId Id     = 3
	AdminName     string = "admin"
	SellerName    string = "seller"
	CashierName   string = "cashier"
)

type RoleId struct {
	Id
}

func NewRoleId(id Id) RoleId {
	if id != AdminRoleId && id != SellerRoleId && id != CashierRoleId {
		panic(fmt.Sprintf("invalid role id: %d", id))
	}

	return RoleId{Id: id}
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

func (roleId RoleId) Name() string {
	switch roleId.Id {
	case AdminRoleId:
		return AdminName
	case SellerRoleId:
		return SellerName
	case CashierRoleId:
		return CashierName
	default:
		panic(fmt.Sprintf("unknown role id: %d", roleId.Id))
	}
}

var UnknownRoleError = errors.New("unknown role")

func ParseRole(role string) (RoleId, error) {
	switch role {
	case "admin":
		return RoleId{Id: AdminRoleId}, nil
	case "seller":
		return RoleId{Id: SellerRoleId}, nil
	case "cashier":
		return RoleId{Id: CashierRoleId}, nil
	default:
		return RoleId{}, fmt.Errorf("unknown role %s: %w", role, UnknownRoleError)
	}
}

func (roleId RoleId) IsAdmin() bool {
	return roleId.Id == AdminRoleId
}

func (roleId RoleId) IsSeller() bool {
	return roleId.Id == SellerRoleId
}

func (roleId RoleId) IsCashier() bool {
	return roleId.Id == CashierRoleId
}

type RoleVisitor[T any] interface {
	Admin() T
	Seller() T
	Cashier() T
}

func VisitRole[T any](roleId RoleId, visitor RoleVisitor[T]) T {
	switch roleId.Id {
	case AdminRoleId:
		return visitor.Admin()
	case SellerRoleId:
		return visitor.Seller()
	case CashierRoleId:
		return visitor.Cashier()
	default:
		panic(fmt.Sprintf("unknown role id: %d", roleId.Id))
	}
}

func (roleId RoleId) IsValid() bool {
	return roleId.Id == AdminRoleId || roleId.Id == SellerRoleId || roleId.Id == CashierRoleId
}
