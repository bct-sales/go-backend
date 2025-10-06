//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
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
			for _, roleID := range models.ListRoles() {
				t.Run(fmt.Sprintf("With role id %d", roleID), func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					userID, err := queries.AddUser(db, roleID, 0, nil, password)
					require.NoError(t, err)

					userExists, err := queries.UserWithIDExists(db, userID)
					require.NoError(t, err)
					require.True(t, userExists)
				})
			}
		}
	})

	t.Run("Fail due to invalid role", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		roleID := models.RoleID{ID: 999} // Assuming this ID does not exist in the database
		password := "xyz"
		createdAt := models.Timestamp(0)
		var lastActivity *models.Timestamp = nil

		_, err := queries.AddUser(db, roleID, createdAt, lastActivity, password)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchRole)
	})
}
