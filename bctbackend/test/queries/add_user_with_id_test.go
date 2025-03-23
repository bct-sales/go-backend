//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAddUserWithId(t *testing.T) {
	for _, password := range []string{"a", "xyz"} {
		for _, userId := range []models.Id{1, 5} {
			for _, roleId := range []models.Id{models.AdminRoleId, models.CashierRoleId, models.SellerRoleId} {
				t.Run(fmt.Sprintf("With role id %d", roleId), func(t *testing.T) {
					db := OpenInitializedDatabase()
					defer db.Close()

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
}

func TestAddUserWithExistingId(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

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
}

func TestAddUserWithIdWithInvalidRole(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

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
}
