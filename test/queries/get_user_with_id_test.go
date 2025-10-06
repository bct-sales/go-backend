//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetUserWithID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		lastActivity := models.Timestamp(2)
		user := models.User{
			Password:     "xyz",
			UserID:       models.ID(1),
			RoleID:       models.NewSellerRoleID(),
			CreatedAt:    models.Timestamp(1),
			LastActivity: &lastActivity,
		}

		user.UserID = setup.User(user.RoleID, aux.WithCreatedAt(user.CreatedAt), aux.WithLastActivity(*user.LastActivity), aux.WithPassword(user.Password)).UserID

		actual, err := queries.GetUserWithID(db, user.UserID)
		require.NoError(t, err)
		require.Equal(t, user, *actual)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		userID := models.ID(999)
		setup.RequireNoSuchUsers(t, userID)

		_, err := queries.GetUserWithID(db, userID)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
	})
}
