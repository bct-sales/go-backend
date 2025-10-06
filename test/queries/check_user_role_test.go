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
	UserId models.ID
	RoleId models.RoleID
}

func TestCheckUserRole(t *testing.T) {
	t.Run("Check correct role", func(t *testing.T) {
		sellerId := models.ID(1)
		adminId := models.ID(2)
		cashierId := models.ID(3)

		for _, pair := range []pair{
			{UserId: sellerId, RoleId: models.NewSellerRoleID()},
			{UserId: adminId, RoleId: models.NewAdminRoleID()},
			{UserId: cashierId, RoleId: models.NewCashierRoleID()},
		} {
			roleName := pair.RoleId.Name()

			t.Run(roleName, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				setup.Cashier(aux.WithUserID(cashierId))
				setup.Admin(aux.WithUserID(adminId))
				setup.Seller(aux.WithUserID(sellerId))

				err := queries.EnsureUserExistsAndHasRole(db, pair.UserId, pair.RoleId)
				require.NoError(t, err)
			})
		}
	})

	t.Run("Check incorrect role", func(t *testing.T) {
		sellerId := models.ID(1)
		adminId := models.ID(2)
		cashierId := models.ID(3)

		for _, pair := range []pair{
			{UserId: adminId, RoleId: models.NewSellerRoleID()},
			{UserId: cashierId, RoleId: models.NewSellerRoleID()},
			{UserId: sellerId, RoleId: models.NewAdminRoleID()},
			{UserId: cashierId, RoleId: models.NewAdminRoleID()},
			{UserId: sellerId, RoleId: models.NewCashierRoleID()},
			{UserId: adminId, RoleId: models.NewCashierRoleID()},
		} {
			roleName := pair.RoleId.Name()

			t.Run(roleName, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				setup.Cashier(aux.WithUserID(cashierId))
				setup.Admin(aux.WithUserID(adminId))
				setup.Seller(aux.WithUserID(sellerId))

				err := queries.EnsureUserExistsAndHasRole(db, pair.UserId, pair.RoleId)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		}
	})

	t.Run("Check non-existing user", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		invalidId := models.ID(9999)

		err := queries.EnsureUserExistsAndHasRole(db, invalidId, models.NewAdminRoleID())
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
	})
}
