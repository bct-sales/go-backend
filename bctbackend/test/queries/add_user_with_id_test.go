//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"testing"

	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAddUserWithId(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, password := range []string{"a", "xyz"} {
			for _, userId := range []models.Id{1, 5} {
				for _, roleId := range []models.Id{models.AdminRoleId, models.CashierRoleId, models.SellerRoleId} {
					t.Run(fmt.Sprintf("With role id %d", roleId), func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						err := queries.AddUserWithId(db, userId, roleId, 0, nil, password)
						require.NoError(t, err)

						userExists, err := queries.UserWithIdExists(db, userId)
						require.NoError(t, err)
						require.True(t, userExists)

						actualRoleId, err := queries.AuthenticateUser(db, userId, password)
						require.NoError(t, err)
						require.Equal(t, roleId, actualRoleId)
					})
				}
			}
		}
	})

	t.Run("Fail due to existing user id", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		userId := models.NewId(1)
		roleId := models.SellerRoleId
		password := "xyz"
		createdAt := models.Timestamp(0)
		var lastAccess *models.Timestamp = nil

		{
			err := queries.AddUserWithId(db, userId, roleId, createdAt, lastAccess, password)
			require.NoError(t, err)
		}

		{
			err := queries.AddUserWithId(db, userId, roleId, createdAt, lastAccess, password)
			var userIdAlreadyInUseError *queries.UserIdAlreadyInUseError
			require.ErrorAs(t, err, &userIdAlreadyInUseError)
		}
	})

	t.Run("Fail due to invalid role id", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		userId := models.NewId(1)
		roleId := models.Id(10)
		password := "xyz"
		createdAt := models.Timestamp(0)
		var lastAccess *models.Timestamp = nil

		require.False(t, models.IsValidRole(roleId), "sanity test: role id should be invalid")

		{
			err := queries.AddUserWithId(db, userId, roleId, createdAt, lastAccess, password)
			var noSuchRoleError *queries.NoSuchRoleError
			require.ErrorAs(t, err, &noSuchRoleError)
		}
	})
}
