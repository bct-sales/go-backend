//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"testing"

	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestAuthentication(t *testing.T) {
	t.Run("Successful authentication", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		password := "xyz"
		userID := models.ID(1)
		roleID := models.NewSellerRoleID()
		createdAt := models.Timestamp(0)
		var lastActivity *models.Timestamp = nil

		queries.AddUserWithID(t.Context(), db, userID, roleID, createdAt, lastActivity, password)

		actualRoleID, err := queries.AuthenticateUser(t.Context(), db, userID, password)
		require.NoError(t, err)
		require.Equal(t, roleID, actualRoleID)
	})

	t.Run("Authenticating non-existing user", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		nonexistentUserID := setup.GenerateNonexistentUserID(t)
		password := "xyz"

		{
			_, err := queries.AuthenticateUser(t.Context(), db, nonexistentUserID, password)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		}
	})

	t.Run("Authenticating using wrong password", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		password := "xyz"
		wrongPassword := "abc"
		userID := models.ID(5)
		roleID := models.NewSellerRoleID()

		queries.AddUserWithID(t.Context(), db, userID, roleID, 0, nil, password)

		_, err := queries.AuthenticateUser(t.Context(), db, userID, wrongPassword)
		requireDatabaseWrappedError(t, err, dberr.ErrWrongPassword)
	})
}
