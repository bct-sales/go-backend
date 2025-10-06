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

func TestAddUserWithID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, password := range []string{"a", "xyz"} {
			for _, userID := range []models.ID{1, 5} {
				for _, roleID := range models.ListRoles() {
					t.Run(fmt.Sprintf("With role id %d", roleID), func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						err := queries.AddUserWithID(db, userID, roleID, 0, nil, password)
						require.NoError(t, err)

						userExists, err := queries.UserWithIDExists(db, userID)
						require.NoError(t, err)
						require.True(t, userExists)

						actualRoleID, err := queries.AuthenticateUser(db, userID, password)
						require.NoError(t, err)
						require.Equal(t, roleID, actualRoleID)
					})
				}
			}
		}
	})

	t.Run("Fail due to existing user id", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		userID := models.ID(1)
		roleID := models.NewSellerRoleID()
		password := "xyz"
		createdAt := models.Timestamp(0)
		var lastAccess *models.Timestamp = nil

		{
			err := queries.AddUserWithID(db, userID, roleID, createdAt, lastAccess, password)
			require.NoError(t, err)
		}

		{
			err := queries.AddUserWithID(db, userID, roleID, createdAt, lastAccess, password)
			requireDatabaseWrappedError(t, err, dberr.ErrIDAlreadyInUse)
		}
	})

	t.Run("Fail due to invalid role id", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		userID := models.ID(1)
		roleID := models.RoleID{ID: 999} // Assuming this ID does not exist in the database
		password := "xyz"
		createdAt := models.Timestamp(0)
		var lastAccess *models.Timestamp = nil

		err := queries.AddUserWithID(db, userID, roleID, createdAt, lastAccess, password)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchRole)
	})
}
