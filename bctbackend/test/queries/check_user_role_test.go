//go:build test

package queries

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

type pair struct {
	UserId models.Id
	RoleId models.Id
}

func TestCheckUserRole(t *testing.T) {
	t.Run("Check correct role", func(t *testing.T) {
		sellerId := models.Id(1)
		adminId := models.Id(2)
		cashierId := models.Id(3)

		for _, pair := range []pair{
			{UserId: sellerId, RoleId: models.SellerRoleId},
			{UserId: adminId, RoleId: models.AdminRoleId},
			{UserId: cashierId, RoleId: models.CashierRoleId},
		} {
			roleName, err := models.NameOfRole(pair.RoleId)
			require.NoError(t, err)

			t.Run(roleName, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				setup.Cashier(aux.WithUserId(cashierId))
				setup.Admin(aux.WithUserId(adminId))
				setup.Seller(aux.WithUserId(sellerId))

				err := queries.EnsureUserExistsAndHasRole(db, pair.UserId, pair.RoleId)
				require.NoError(t, err)
			})
		}
	})

	t.Run("Check incorrect role", func(t *testing.T) {
		sellerId := models.Id(1)
		adminId := models.Id(2)
		cashierId := models.Id(3)

		for _, pair := range []pair{
			{UserId: adminId, RoleId: models.SellerRoleId},
			{UserId: cashierId, RoleId: models.SellerRoleId},
			{UserId: sellerId, RoleId: models.AdminRoleId},
			{UserId: cashierId, RoleId: models.AdminRoleId},
			{UserId: sellerId, RoleId: models.CashierRoleId},
			{UserId: adminId, RoleId: models.CashierRoleId},
		} {
			roleName, err := models.NameOfRole(pair.RoleId)
			require.NoError(t, err)

			t.Run(roleName, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				setup.Cashier(aux.WithUserId(cashierId))
				setup.Admin(aux.WithUserId(adminId))
				setup.Seller(aux.WithUserId(sellerId))

				err := queries.EnsureUserExistsAndHasRole(db, pair.UserId, pair.RoleId)
				require.ErrorIs(t, err, database.ErrWrongRole)
			})
		}
	})

	t.Run("Check non-existing user", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		invalidId := models.Id(9999)

		err := queries.EnsureUserExistsAndHasRole(db, invalidId, models.AdminRoleId)
		require.ErrorIs(t, err, database.ErrNoSuchUser)
	})
}
