package queries

import (
	dberr "bctbackend/database/errors"
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
func AuthenticateUser(db *sql.DB, userId models.Id, password string) (models.RoleId, error) {
	row := db.QueryRow(
		`
			SELECT role_id, password
			FROM users
			where user_id = $1
		`,
		userId)

	var roleId models.RoleId
	var expectedPassword string
	err := row.Scan(&roleId.Id, &expectedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.RoleId{}, fmt.Errorf("failed to authenticate user %d: %w", userId, dberr.ErrNoSuchUser)
		}

		return models.RoleId{}, fmt.Errorf("failed to execute query to look up user %d in database: %w", userId, err)
	}

	if expectedPassword != password {
		return models.RoleId{}, dberr.ErrWrongPassword
	}

	return roleId, nil
}
