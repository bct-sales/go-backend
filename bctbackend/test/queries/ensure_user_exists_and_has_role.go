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
	_ "modernc.org/sqlite"
)

func TestEnsureUserExistsAndHasRole(t *testing.T) {
	roleIds := []models.Id{
		models.SellerRoleId,
		models.CashierRoleId,
		models.AdminRoleId,
	}

	t.Run("Success", func(t *testing.T) {
		for _, roleId := range roleIds {
			testLabel := roleId.String()

			t.Run(testLabel, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				user := setup.User(roleId)

				err := queries.EnsureUserExistsAndHasRole(db, user.UserId, roleId)
				require.NoError(t, err)
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Wrong role", func(t *testing.T) {
			for _, expectedRoleId := range roleIds {
				for _, actualRoleId := range roleIds {
					if expectedRoleId != actualRoleId {
						testLabel := fmt.Sprintf("Expected role: %s, actual role: %s", expectedRoleId, actualRoleId)
						t.Run(testLabel, func(t *testing.T) {
							setup, db := NewDatabaseFixture(WithDefaultCategories)
							defer setup.Close()

							user := setup.User(actualRoleId)

							err := queries.EnsureUserExistsAndHasRole(db, user.UserId, expectedRoleId)
							require.ErrorIs(t, err, database.ErrWrongRole)
						})
					}
				}
			}
		})

		t.Run("User does not exist", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			nonexistentUserId := models.Id(9999) // Assuming this ID does not exist in the database
			setup.RequireNoSuchUsers(t, nonexistentUserId)

			err := queries.EnsureUserExistsAndHasRole(db, nonexistentUserId, models.SellerRoleId)
			require.ErrorIs(t, err, database.ErrNoSuchUser)
		})

		t.Run("Role does not exist", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			nonexistentRole := models.Id(9999)

			err := queries.EnsureUserExistsAndHasRole(db, seller.UserId, nonexistentRole)
			require.ErrorIs(t, err, database.ErrNoSuchRole)
		})
	})
}
