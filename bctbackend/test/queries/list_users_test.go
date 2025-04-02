//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetUsers(t *testing.T) {
	t.Run("Single user", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		password := "xyz"
		userId := models.NewId(1)
		roleId := models.SellerRoleId
		createdAt := models.Timestamp(1)
		lastActivity := models.Timestamp(2)

		queries.AddUserWithId(db, userId, roleId, createdAt, &lastActivity, password)

		users := []*models.User{}
		err := queries.GetUsers(db, queries.CollectTo(&users))
		require.NoError(t, err)
		require.Len(t, users, 1)
		require.Equal(t, userId, users[0].UserId)
		require.Equal(t, roleId, users[0].RoleId)
		require.Equal(t, password, users[0].Password)
		require.Equal(t, createdAt, users[0].CreatedAt)
		require.NotNil(t, users[0].LastActivity)
		require.Equal(t, lastActivity, *users[0].LastActivity)
	})

	t.Run("Two users", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		user1 := models.User{
			UserId:       models.NewId(1),
			RoleId:       models.SellerRoleId,
			CreatedAt:    models.Timestamp(1),
			LastActivity: nil,
			Password:     "xyz",
		}

		lastActivity2 := models.Timestamp(50)
		user2 := models.User{
			UserId:       models.NewId(2),
			RoleId:       models.AdminRoleId,
			CreatedAt:    models.Timestamp(2),
			LastActivity: &lastActivity2,
			Password:     "abc",
		}

		queries.AddUserWithId(db, user1.UserId, user1.RoleId, user1.CreatedAt, user1.LastActivity, user1.Password)
		queries.AddUserWithId(db, user2.UserId, user2.RoleId, user2.CreatedAt, user2.LastActivity, user2.Password)

		users := []*models.User{}
		err := queries.GetUsers(db, queries.CollectTo(&users))
		require.NoError(t, err)
		require.Len(t, users, 2)
		require.Equal(t, user1, *users[0])
		require.Equal(t, user2, *users[1])
	})
}
