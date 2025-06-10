//go:build test

package queries

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, password := range []string{"a", "xyz"} {
			for _, roleId := range []models.RoleId{models.NewAdminRoleId(), models.NewCashierRoleId(), models.NewSellerRoleId()} {
				t.Run(fmt.Sprintf("With role id %d", roleId), func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					userId, err := queries.AddUser(db, roleId, 0, nil, password)
					require.NoError(t, err)

					userExists, err := queries.UserWithIdExists(db, userId)
					require.NoError(t, err)
					require.True(t, userExists)
				})
			}
		}
	})

	t.Run("Fail due to invalid role", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		roleId := models.RoleId{Id: 999} // Assuming this ID does not exist in the database
		password := "xyz"
		createdAt := models.Timestamp(0)
		var lastActivity *models.Timestamp = nil

		_, err := queries.AddUser(db, roleId, createdAt, lastActivity, password)
		require.ErrorIs(t, err, database.ErrNoSuchRole)
	})
}
