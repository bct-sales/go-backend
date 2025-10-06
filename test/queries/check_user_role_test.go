//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

type pair struct {
	UserID models.ID
	RoleID models.RoleID
}

func TestCheckUserRole(t *testing.T) {
	t.Run("Check correct role", func(t *testing.T) {
		sellerID := models.ID(1)
		adminID := models.ID(2)
		cashierID := models.ID(3)

		for _, pair := range []pair{
			{UserID: sellerID, RoleID: models.NewSellerRoleID()},
			{UserID: adminID, RoleID: models.NewAdminRoleID()},
			{UserID: cashierID, RoleID: models.NewCashierRoleID()},
		} {
			roleName := pair.RoleID.Name()

			t.Run(roleName, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				setup.Cashier(aux.WithUserID(cashierID))
				setup.Admin(aux.WithUserID(adminID))
				setup.Seller(aux.WithUserID(sellerID))

				err := queries.EnsureUserExistsAndHasRole(db, pair.UserID, pair.RoleID)
				require.NoError(t, err)
			})
		}
	})

	t.Run("Check incorrect role", func(t *testing.T) {
		sellerID := models.ID(1)
		adminID := models.ID(2)
		cashierID := models.ID(3)

		for _, pair := range []pair{
			{UserID: adminID, RoleID: models.NewSellerRoleID()},
			{UserID: cashierID, RoleID: models.NewSellerRoleID()},
			{UserID: sellerID, RoleID: models.NewAdminRoleID()},
			{UserID: cashierID, RoleID: models.NewAdminRoleID()},
			{UserID: sellerID, RoleID: models.NewCashierRoleID()},
			{UserID: adminID, RoleID: models.NewCashierRoleID()},
		} {
			roleName := pair.RoleID.Name()

			t.Run(roleName, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				setup.Cashier(aux.WithUserID(cashierID))
				setup.Admin(aux.WithUserID(adminID))
				setup.Seller(aux.WithUserID(sellerID))

				err := queries.EnsureUserExistsAndHasRole(db, pair.UserID, pair.RoleID)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		}
	})

	t.Run("Check non-existing user", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		invalidID := models.ID(9999)

		err := queries.EnsureUserExistsAndHasRole(db, invalidID, models.NewAdminRoleID())
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
	})
}
