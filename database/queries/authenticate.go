package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/meta"
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
)

// AuthenticateUser authenticates a user with the given user id and password.
// If the user is authenticated, the function returns the role id of the user.
// If the user is not authenticated, the function returns an error.
// If the user does not exist, the function returns an ErrNoSuchUser.
// If the password is wrong, the function returns a ErrWrongPassword.
// If there is an error while querying the database, the function returns the error.
func AuthenticateUser(database DatabaseQuerier, userII models.ID, password string) (r_result models.RoleID, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query := squirrel.Select(meta.User.RoleID, meta.User.Password).From(meta.User.Table).Where(squirrel.Eq{meta.User.UserID: userII})
	row := query.RunWith(database).QueryRow()

	var roleID models.RoleID
	var expectedPassword string
	err := row.Scan(&roleID.ID, &expectedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.RoleID{}, fmt.Errorf("failed to authenticate user %d: %w", userII, dberr.ErrNoSuchUser)
		}

		return models.RoleID{}, fmt.Errorf("failed to execute query to look up user %d in database: %w", userII, err)
	}

	if expectedPassword != password {
		return models.RoleID{}, dberr.ErrWrongPassword
	}

	return roleID, nil
}
