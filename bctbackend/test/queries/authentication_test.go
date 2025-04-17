//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"testing"

	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAuthentication(t *testing.T) {
	t.Run("Successful authentication", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		password := "xyz"
		userId := models.NewId(1)
		roleId := models.SellerRoleId
		createdAt := models.NewTimestamp(0)
		var lastActivity *models.Timestamp = nil

		queries.AddUserWithId(db, userId, roleId, createdAt, lastActivity, password)

		actualRoleId, err := queries.AuthenticateUser(db, userId, password)
		require.NoError(t, err)
		require.Equal(t, roleId, actualRoleId)
	})

	t.Run("Authenticating non-existing user", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		password := "xyz"
		userId := models.NewId(5)

		{
			userExists, err := queries.UserWithIdExists(db, userId)
			require.NoError(t, err)
			require.False(t, userExists)
		}

		{
			_, err := queries.AuthenticateUser(db, userId, password)
			var noSuchUserError *queries.NoSuchUserError
			require.ErrorAs(t, err, &noSuchUserError)
		}
	})

	t.Run("Authenticating using wrong password", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		password := "xyz"
		wrongPassword := "abc"
		userId := models.NewId(5)
		roleId := models.SellerRoleId

		queries.AddUserWithId(db, userId, roleId, 0, nil, password)

		_, err := queries.AuthenticateUser(db, userId, wrongPassword)
		var wrongPasswordError *queries.WrongPasswordError
		require.ErrorAs(t, err, &wrongPasswordError)
	})
}
