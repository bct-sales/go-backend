//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetSellerItemCount(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Single seller", func(t *testing.T) {
			for _, itemCount := range []int64{0, 1, 2, 10, 100} {
				testLabel := fmt.Sprintf("Seller with %d items", itemCount)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture()
					defer setup.Close()

					seller := setup.Seller()

					for i := int64(0); i < itemCount; i++ {
						setup.Item(seller.UserId, aux.WithDummyData(int(i)))
					}

					actual, err := queries.GetSellerItemCount(db, seller.UserId)
					require.NoError(t, err)
					require.Equal(t, itemCount, actual)
				})
			}
		})

		t.Run("Multiple sellers", func(t *testing.T) {
			for _, itemCount := range []int64{0, 1, 2, 10, 100} {
				testLabel := fmt.Sprintf("Seller with %d items", itemCount)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture()
					defer setup.Close()

					seller := setup.Seller()
					otherSeller := setup.Seller()

					for i := int64(0); i < itemCount; i++ {
						setup.Item(seller.UserId, aux.WithDummyData(int(i)))
						setup.Item(otherSeller.UserId, aux.WithDummyData(int(3*i)))
					}

					actual, err := queries.GetSellerItemCount(db, seller.UserId)
					require.NoError(t, err)
					require.Equal(t, itemCount, actual)
				})
			}
		})

		t.Run("Frozen items are counted", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()

			itemCount := int64(10)
			for i := int64(0); i < itemCount; i++ {
				setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithFrozen(true))
			}

			actual, err := queries.GetSellerItemCount(db, seller.UserId)
			require.NoError(t, err)
			require.Equal(t, itemCount, actual)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No such user", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			nonExistentSellerId := models.Id(1000)
			setup.RequireNoSuchUser(t, nonExistentSellerId)

			_, err := queries.GetSellerItemCount(db, nonExistentSellerId)
			{
				var noSuchUserError *queries.NoSuchUserError
				require.ErrorAs(t, err, &noSuchUserError)
			}
		})

		t.Run("List items of cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			cashier := setup.Cashier()

			_, err := queries.GetSellerItemCount(db, cashier.UserId)
			{
				var invalidRoleError *queries.InvalidRoleError
				require.ErrorAs(t, err, &invalidRoleError)
			}
		})

		t.Run("List items of admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			admin := setup.Admin()

			_, err := queries.GetSellerItemCount(db, admin.UserId)
			{
				var invalidRoleError *queries.InvalidRoleError
				require.ErrorAs(t, err, &invalidRoleError)
			}
		})
	})
}
