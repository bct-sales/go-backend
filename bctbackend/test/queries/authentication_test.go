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
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		password := "xyz"
		userId := models.Id(1)
		roleId := models.SellerRoleId
		createdAt := models.Timestamp(0)
		var lastActivity *models.Timestamp = nil

		queries.AddUserWithId(db, userId, roleId, createdAt, lastActivity, password)

		actualRoleId, err := queries.AuthenticateUser(db, userId, password)
		require.NoError(t, err)
		require.Equal(t, roleId, actualRoleId)
	})

	t.Run("Authenticating non-existing user", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		userId := models.Id(5)
		password := "xyz"

		setup.RequireNoSuchUsers(t, userId)

		{
			_, err := queries.AuthenticateUser(db, userId, password)
			require.ErrorIs(t, err, queries.ErrNoSuchUser)
		}
	})

	t.Run("Authenticating using wrong password", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		password := "xyz"
		wrongPassword := "abc"
		userId := models.Id(5)
		roleId := models.SellerRoleId

		queries.AddUserWithId(db, userId, roleId, 0, nil, password)

		_, err := queries.AuthenticateUser(db, userId, wrongPassword)
		require.ErrorIs(t, err, queries.ErrWrongPassword)
	})
}
