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

func TestGetUserWithId(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		lastActivity := models.Timestamp(2)
		user := models.User{
			Password:     "xyz",
			UserId:       models.ID(1),
			RoleId:       models.NewSellerRoleId(),
			CreatedAt:    models.Timestamp(1),
			LastActivity: &lastActivity,
		}

		user.UserId = setup.User(user.RoleId, aux.WithCreatedAt(user.CreatedAt), aux.WithLastActivity(*user.LastActivity), aux.WithPassword(user.Password)).UserId

		actual, err := queries.GetUserWithId(db, user.UserId)
		require.NoError(t, err)
		require.Equal(t, user, *actual)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		userId := models.ID(999)
		setup.RequireNoSuchUsers(t, userId)

		_, err := queries.GetUserWithId(db, userId)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
	})
}
