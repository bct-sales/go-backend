//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAuthentication(t *testing.T) {
	t.Run("Successful authentication", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

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
		db := OpenInitializedDatabase()
		defer db.Close()

		password := "xyz"
		userId := models.NewId(5)

		{
			userExists, err := queries.UserWithIdExists(db, userId)
			require.NoError(t, err)
			require.False(t, userExists)
		}

		{
			_, err := queries.AuthenticateUser(db, userId, password)
			require.Error(t, err)
			require.IsType(t, &queries.NoSuchUserError{}, err)
		}
	})

	t.Run("Authenticating using wrong password", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		password := "xyz"
		wrongPassword := "abc"
		userId := models.NewId(5)
		roleId := models.SellerRoleId

		queries.AddUserWithId(db, userId, roleId, 0, nil, password)

		_, err := queries.AuthenticateUser(db, userId, wrongPassword)
		require.Error(t, err)
		require.IsType(t, &queries.WrongPasswordError{}, err)
	})
}
