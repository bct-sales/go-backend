//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetUsers(t *testing.T) {
	t.Run("Single user", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		password := "xyz"
		userID := models.ID(1)
		roleID := models.NewSellerRoleID()
		createdAt := models.Timestamp(1)
		lastActivity := models.Timestamp(2)

		queries.AddUserWithID(db, userID, roleID, createdAt, &lastActivity, password)

		users := []*models.User{}
		err := queries.GetUsers(db, queries.CollectTo(&users))
		require.NoError(t, err)
		require.Len(t, users, 1)
		require.Equal(t, userID, users[0].UserID)
		require.Equal(t, roleID, users[0].RoleID)
		require.Equal(t, password, users[0].Password)
		require.Equal(t, createdAt, users[0].CreatedAt)
		require.NotNil(t, users[0].LastActivity)
		require.Equal(t, lastActivity, *users[0].LastActivity)
	})

	t.Run("Two users", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		user1 := models.User{
			UserID:       models.ID(1),
			RoleID:       models.NewSellerRoleID(),
			CreatedAt:    models.Timestamp(1),
			LastActivity: nil,
			Password:     "xyz",
		}

		lastActivity2 := models.Timestamp(50)
		user2 := models.User{
			UserID:       models.ID(2),
			RoleID:       models.NewAdminRoleID(),
			CreatedAt:    models.Timestamp(2),
			LastActivity: &lastActivity2,
			Password:     "abc",
		}

		queries.AddUserWithID(db, user1.UserID, user1.RoleID, user1.CreatedAt, user1.LastActivity, user1.Password)
		queries.AddUserWithID(db, user2.UserID, user2.RoleID, user2.CreatedAt, user2.LastActivity, user2.Password)

		users := []*models.User{}
		err := queries.GetUsers(db, queries.CollectTo(&users))
		require.NoError(t, err)
		require.Len(t, users, 2)
		require.Equal(t, user1, *users[0])
		require.Equal(t, user2, *users[1])
	})
}
