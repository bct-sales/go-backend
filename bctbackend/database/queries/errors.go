package queries

import (
	models "bctbackend/database/models"
	"errors"
	"fmt"
)

var ErrUserIdAlreadyInUse = errors.New("user id already in use")
var ErrNoSuchItem = errors.New("no such item")
var ErrSaleMissingItems = errors.New("sale must have at least one item")
var ErrSaleRequiresCashier = errors.New("sale requires a cashier")
var InvalidRoleError = errors.New("user has an invalid role")
var NoSuchUserError = errors.New("no such user")

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

var ItemFrozenError = errors.New("item is frozen")
var ItemHiddenError = errors.New("item is hidden")
var InvalidCategoryNameError = errors.New("category name is invalid")

type CategoryIdAlreadyInUseError struct{}

type HiddenFrozenItemError struct{}

func (err *NoSuchSaleError) Error() string {
	return fmt.Sprintf("no sale with id %v", err.SaleId)
}

func (err *NoSuchCategoryError) Error() string {
	return fmt.Sprintf("no category with id %v", err.CategoryId)
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

func (e *InvalidItemDescriptionError) Error() string {
	return fmt.Sprintf("invalid description \"%s\"", e.Description)
}

func (e *CategoryIdAlreadyInUseError) Error() string {
	return "category id already in use"
}

func (e *HiddenFrozenItemError) Error() string {
	return "items cannot be hidden and frozen at the same time"
}
