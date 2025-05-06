package queries

import (
	models "bctbackend/database/models"
	"fmt"
)

type UserIdAlreadyInUseError struct {
	UserId models.Id
}

type NoSuchItemError struct {
	Id models.Id
}

type SaleMissingItemsError struct{}

type SaleRequiresCashierError struct{}

type InvalidRoleError struct {
	UserId         models.Id
	ExpectedRoleId models.Id
}

type NoSuchUserError struct {
	UserId models.Id
}

type WrongPasswordError struct{}

type NoSessionFoundError struct{}

type NoSuchSaleError struct {
	SaleId models.Id
}

type NoSuchSessionError struct {
	SessionId models.SessionId
}

type NoSuchCategoryError struct {
	CategoryId models.Id
}

type NoSuchRoleError struct {
	RoleId models.Id
}

type InvalidPriceError struct {
	PriceInCents models.MoneyInCents
}

type InvalidItemDescriptionError struct {
	Description string
}

type DuplicateItemInSaleError struct {
	ItemId models.Id
}

type ItemFrozenError struct {
	Id models.Id
}

func (err *UserIdAlreadyInUseError) Error() string {
	return fmt.Sprintf("user id %d already in use", err.UserId)
}

func (err *NoSuchSaleError) Error() string {
	return fmt.Sprintf("no sale with id %v", err.SaleId)
}

func (err *NoSuchCategoryError) Error() string {
	return fmt.Sprintf("no category with id %v", err.CategoryId)
}

func (e *NoSuchItemError) Error() string {
	return fmt.Sprintf("item with id %d not found", e.Id)
}

func (e *NoSuchItemError) Unwrap() error {
	return nil
}

func (e *SaleMissingItemsError) Error() string {
	return "sale must have at least one item"
}

func (e *SaleRequiresCashierError) Error() string {
	return "sale requires a cashier"
}

func (e *NoSuchUserError) Error() string {
	return fmt.Sprintf("unknown user %d", e.UserId)
}

func (e *WrongPasswordError) Error() string {
	return "wrong password"
}

func (e *NoSessionFoundError) Error() string {
	return "no session found"
}

func (e *NoSuchSessionError) Error() string {
	return fmt.Sprintf("no session with id %v", e.SessionId)
}

func (e *NoSuchRoleError) Error() string {
	return fmt.Sprintf("no role with id %v", e.RoleId)
}

func (e *InvalidPriceError) Error() string {
	return fmt.Sprintf("price %d is invalid", e.PriceInCents)
}

func (e *DuplicateItemInSaleError) Error() string {
	return fmt.Sprintf("item %d is duplicated in sale", e.ItemId)
}

func (e *InvalidRoleError) Error() string {
	return fmt.Sprintf("user %d should have role %d", e.UserId, e.ExpectedRoleId)
}

func (e *ItemFrozenError) Error() string {
	return fmt.Sprintf("item %d is frozen", e.Id)
}

func (e *InvalidItemDescriptionError) Error() string {
	return fmt.Sprintf("invalid description \"%s\"", e.Description)
}
