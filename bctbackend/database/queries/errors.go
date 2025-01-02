package queries

import (
	models "bctbackend/database/models"
	"fmt"
)

type ItemNotFoundError struct {
	Id models.Id
}

type SaleMissingItemsError struct{}

type ItemRequiresSellerError struct{}

type SaleRequiresCashierError struct{}

type UnknownUserError struct {
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

func (err NoSuchSaleError) Error() string {
	return fmt.Sprintf("no sale with id %v", err.SaleId)
}

func (e *ItemNotFoundError) Error() string {
	return fmt.Sprintf("item with id %d not found", e.Id)
}

func (e *ItemNotFoundError) Unwrap() error {
	return nil
}

func (e *SaleMissingItemsError) Error() string {
	return "sale must have at least one item"
}

func (e *ItemRequiresSellerError) Error() string {
	return "item needs seller as owner"
}

func (e *SaleRequiresCashierError) Error() string {
	return "sale requires a cashier"
}

func (e *UnknownUserError) Error() string {
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
