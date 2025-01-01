package errors

import (
	"bctbackend/database/models"
	"fmt"
)

type ItemNotFoundError struct {
	Id models.Id
}

func (e *ItemNotFoundError) Error() string {
	return fmt.Sprintf("item with id %d not found", e.Id)
}

func (e *ItemNotFoundError) Unwrap() error {
	return nil
}
