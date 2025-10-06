//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"testing"

	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestAddUserWithId(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, password := range []string{"a", "xyz"} {
			for _, userId := range []models.ID{1, 5} {
				for _, roleId := range models.ListRoles() {
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

		userId := models.ID(1)
		roleId := models.NewSellerRoleID()
		password := "xyz"
		createdAt := models.Timestamp(0)
		var lastAccess *models.Timestamp = nil

		{
			err := queries.AddUserWithId(db, userId, roleId, createdAt, lastAccess, password)
			require.NoError(t, err)
		}

		{
			err := queries.AddUserWithId(db, userId, roleId, createdAt, lastAccess, password)
			requireDatabaseWrappedError(t, err, dberr.ErrIdAlreadyInUse)
		}
	})

	t.Run("Fail due to invalid role id", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		userId := models.ID(1)
		roleId := models.RoleID{ID: 999} // Assuming this ID does not exist in the database
		password := "xyz"
		createdAt := models.Timestamp(0)
		var lastAccess *models.Timestamp = nil

		err := queries.AddUserWithId(db, userId, roleId, createdAt, lastAccess, password)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchRole)
	})
}
