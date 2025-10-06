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

func TestEnsureUserExistsAndHasRole(t *testing.T) {
	roleIDs := []models.RoleID{
		models.NewSellerRoleID(),
		models.NewCashierRoleID(),
		models.NewAdminRoleID(),
	}

	t.Run("Success", func(t *testing.T) {
		for _, roleID := range roleIDs {
			testLabel := roleID.String()

			t.Run(testLabel, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				user := setup.User(roleID)

				err := queries.EnsureUserExistsAndHasRole(db, user.UserID, roleID)
				require.NoError(t, err)
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Wrong role", func(t *testing.T) {
			for _, expectedRoleID := range roleIDs {
				for _, actualRoleID := range roleIDs {
					if expectedRoleID != actualRoleID {
						testLabel := fmt.Sprintf("Expected role: %s, actual role: %s", expectedRoleID, actualRoleID)
						t.Run(testLabel, func(t *testing.T) {
							setup, db := NewDatabaseFixture(WithDefaultCategories)
							defer setup.Close()

							user := setup.User(actualRoleID)

							err := queries.EnsureUserExistsAndHasRole(db, user.UserID, expectedRoleID)
							requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
						})
					}
				}
			}
		})

		t.Run("User does not exist", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			nonexistentUserID := models.ID(9999) // Assuming this ID does not exist in the database
			setup.RequireNoSuchUsers(t, nonexistentUserID)

			err := queries.EnsureUserExistsAndHasRole(db, nonexistentUserID, models.NewSellerRoleID())
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})
	})
}
