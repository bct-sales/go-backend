package queries

import (
	models "bctbackend/database/models"
	"fmt"
)

type AuthenticationError struct {
	Reason error
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication error: %v", e.Reason)
}

func (e *AuthenticationError) Unwrap() error {
	return e.Reason
}

type ItemNotFoundError struct {
	Id models.Id
}

func (e *ItemNotFoundError) Error() string {
	return fmt.Sprintf("item with id %d not found", e.Id)
}

func (e *ItemNotFoundError) Unwrap() error {
	return nil
}

type SaleMissingItemsError struct{}

func (e *SaleMissingItemsError) Error() string {
	return "sale must have at least one item"
}

type SaleRequiresCashierError struct{}

func (e *SaleRequiresCashierError) Error() string {
	return "sale requires a cashier"
}

type UnknownUserError struct{}

func (e *UnknownUserError) Error() string {
	return "unknown user"
}

type WrongPasswordError struct{}

func (e *WrongPasswordError) Error() string {
	return "wrong password"
}
