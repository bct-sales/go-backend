package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
)

type AuthenticationError struct {
	reason error
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication error: %v", e.reason)
}

func (e *AuthenticationError) Unwrap() error {
	return e.reason
}

type UnknownUserError struct{}

func (e *UnknownUserError) Error() string {
	return "unknown user"
}

type WrongPasswordError struct{}

func (e *WrongPasswordError) Error() string {
	return "wrong password"
}

func AuthenticateUser(db *sql.DB, userId models.Id, password string) (models.Id, error) {
	row := db.QueryRow(
		`
			SELECT role_id, password
			FROM users
			where user_id = $1
		`,
		userId)

	var roleId models.Id
	var expectedPassword string
	err := row.Scan(&roleId, &expectedPassword)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, &AuthenticationError{reason: &UnknownUserError{}}
		}

		return 0, &AuthenticationError{reason: err}
	}

	if expectedPassword == password {
		return roleId, nil
	} else {
		return 0, &AuthenticationError{reason: &WrongPasswordError{}}
	}
}
