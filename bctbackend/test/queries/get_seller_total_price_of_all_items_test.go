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

func TestSellerTotalPriceOfAllTimes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Single seller", func(t *testing.T) {
			for _, itemCount := range []int64{0, 1, 2, 10, 100} {
				testLabel := fmt.Sprintf("Seller with %d items", itemCount)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture()
					defer setup.Close()

					seller := setup.Seller()

					expectedTotal := models.MoneyInCents(0)
					for i := int64(0); i < itemCount; i++ {
						price := models.MoneyInCents((i + 1) * 50)
						setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithPriceInCents(price))
						expectedTotal += price
					}

					actualTotal, err := queries.GetSellerTotalPriceOfAllItems(db, seller.UserId)
					require.NoError(t, err)
					require.Equal(t, expectedTotal, actualTotal)
				})
			}
		})

		t.Run("Multiple seller", func(t *testing.T) {
			for _, itemCount := range []int64{0, 1, 2, 10, 100} {
				testLabel := fmt.Sprintf("Seller with %d items", itemCount)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture()
					defer setup.Close()

					seller := setup.Seller()
					otherSeller := setup.Seller()

					expectedTotal := models.MoneyInCents(0)
					for i := int64(0); i < itemCount; i++ {
						price := models.MoneyInCents((i + 1) * 50)
						setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithPriceInCents(price))
						setup.Item(otherSeller.UserId, aux.WithDummyData(int(i)), aux.WithPriceInCents(price))
						expectedTotal += price
					}

					actualTotal, err := queries.GetSellerTotalPriceOfAllItems(db, seller.UserId)
					require.NoError(t, err)
					require.Equal(t, expectedTotal, actualTotal)
				})
			}
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No such user", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			nonExistentSellerId := models.Id(1000)
			setup.RequireNoSuchUser(t, nonExistentSellerId)

			_, err := queries.GetSellerTotalPriceOfAllItems(db, nonExistentSellerId)
			{
				var noSuchUserError *queries.NoSuchUserError
				require.ErrorAs(t, err, &noSuchUserError)
			}
		})

		t.Run("Sum of item prices of cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			cashier := setup.Cashier()

			_, err := queries.GetSellerTotalPriceOfAllItems(db, cashier.UserId)
			{
				var invalidRoleError *queries.InvalidRoleError
				require.ErrorAs(t, err, &invalidRoleError)
			}
		})

		t.Run("Sum of item prices of admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			admin := setup.Admin()

			_, err := queries.GetSellerTotalPriceOfAllItems(db, admin.UserId)
			{
				var invalidRoleError *queries.InvalidRoleError
				require.ErrorAs(t, err, &invalidRoleError)
			}
		})
	})
}
