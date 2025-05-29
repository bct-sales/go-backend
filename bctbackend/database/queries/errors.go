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
var ErrInvalidRole = errors.New("user has an invalid role")
var ErrNoSuchUser = errors.New("no such user")
var ErrWrongPassword = errors.New("wrong password")
var ErrNoSessionFound = errors.New("no session found")
var ErrNoSuchSale = errors.New("no such sale")
var NoSuchSessionError = errors.New("no such session")
var NoSuchCategoryError = errors.New("no such category")
var NoSuchRoleError = errors.New("no such role")

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
