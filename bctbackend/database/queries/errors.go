package queries

import (
	models "bctbackend/database/models"
	"fmt"
)

type AuthenticationError struct {
	Reason error
}

type ItemNotFoundError struct {
	Id models.Id
}

type SaleMissingItemsError struct{}

type SaleRequiresCashierError struct{}

type UnknownUserError struct {
	UserId models.Id
}

type WrongPasswordError struct{}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication error: %v", e.Reason)
}

func (e *AuthenticationError) Unwrap() error {
	return e.Reason
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

func (e *SaleRequiresCashierError) Error() string {
	return "sale requires a cashier"
}

func (e *UnknownUserError) Error() string {
	return fmt.Sprintf("unknown user %d", e.UserId)
}

func (e *WrongPasswordError) Error() string {
	return "wrong password"
}
