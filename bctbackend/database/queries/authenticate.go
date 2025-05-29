package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
)

// AuthenticateUser authenticates a user with the given user id and password.
// If the user is authenticated, the function returns the role id of the user.
// If the user is not authenticated, the function returns an error.
// If the user does not exist, the function returns an NoSuchUserError.
// If the password is wrong, the function returns a WrongPasswordError.
// If there is an error while querying the database, the function returns the error.
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
			return 0, fmt.Errorf("failed to authenticate user %d: %w", userId, ErrNoSuchUser)
		}

		return 0, err
	}

	if expectedPassword != password {
		return 0, &WrongPasswordError{}
	}

	return roleId, nil
}
