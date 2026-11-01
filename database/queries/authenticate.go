package queries

import (
	"context"
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"bctbackend/database/queries/internal/isql"
	"database/sql"
	"errors"
	"fmt"
)

// AuthenticateUser authenticates a user with the given user id and password.
// If the user is authenticated, the function returns the role id of the user.
// If the user is not authenticated, the function returns an error.
// If the user does not exist, the function returns an ErrNoSuchUser.
// If the password is wrong, the function returns a ErrWrongPassword.
// If there is an error while querying the database, the function returns the error.
func AuthenticateUser(ctx context.Context, database DatabaseQuerier, userID models.ID, password string) (r_result models.RoleID, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row, err := isql.New(database).AuthenticateUser(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.RoleID{}, fmt.Errorf("failed to authenticate user %d: %w", userID, dberr.ErrNoSuchUser)
		}

		return models.RoleID{}, fmt.Errorf("failed to execute query to look up user %d in database: %w", userID, err)
	}

	if row.Password != password {
		return models.RoleID{}, dberr.ErrWrongPassword
	}

	return models.RoleID{ID: models.ID(row.RoleID)}, nil
}
