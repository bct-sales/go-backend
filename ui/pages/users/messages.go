package users

import (
	"bctbackend/database/models"
)

type databaseErrorMessage struct {
	message string
	err     error
}

type usersFetchedMessage struct {
	users []*models.User
}
