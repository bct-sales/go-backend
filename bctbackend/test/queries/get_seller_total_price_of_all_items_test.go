//go:build test

package queries

import (
	"bctbackend/database"
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
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()

					expectedTotal := models.MoneyInCents(0)
					for i := int64(0); i < itemCount; i++ {
						price := models.MoneyInCents((i + 1) * 50)
						setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithHidden(false))
						expectedTotal += price
					}

					actualTotal, err := queries.GetSellerTotalPriceOfAllItems(db, seller.UserId, queries.AllItems)
					require.NoError(t, err)
					require.Equal(t, expectedTotal, actualTotal)
				})
			}
		})

		t.Run("Multiple seller", func(t *testing.T) {
			for _, itemCount := range []int64{0, 1, 2, 10, 100} {
				testLabel := fmt.Sprintf("Seller with %d items", itemCount)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()
					otherSeller := setup.Seller()

					expectedTotal := models.MoneyInCents(0)
					for i := int64(0); i < itemCount; i++ {
						price := models.MoneyInCents((i + 1) * 50)
						setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithHidden(false))
						setup.Item(otherSeller.UserId, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithHidden(false))
						expectedTotal += price
					}

					actualTotal, err := queries.GetSellerTotalPriceOfAllItems(db, seller.UserId, queries.AllItems)
					require.NoError(t, err)
					require.Equal(t, expectedTotal, actualTotal)
				})
			}
		})

		t.Run("Count all items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			expectedTotal := models.MoneyInCents(0)

			// Add visible items
			for i := 0; i < 10; i++ {
				price := models.MoneyInCents((i + 1) * 50)
				setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithFrozen(false), aux.WithHidden(false))
				expectedTotal += price
			}

			// Add hidden items
			for i := 0; i < 10; i++ {
				price := models.MoneyInCents((i + 1) * 150)
				setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithFrozen(false), aux.WithHidden(true))
				expectedTotal += price
			}

			actualTotal, err := queries.GetSellerTotalPriceOfAllItems(db, seller.UserId, queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, expectedTotal, actualTotal)
		})

		t.Run("Count only hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			expectedTotal := models.MoneyInCents(0)

			// Add visible items
			for i := 0; i < 10; i++ {
				price := models.MoneyInCents((i + 1) * 50)
				setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithFrozen(false), aux.WithHidden(false))
			}

			// Add hidden items
			for i := 0; i < 10; i++ {
				price := models.MoneyInCents((i + 1) * 150)
				setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithFrozen(false), aux.WithHidden(true))
				expectedTotal += price
			}

			actualTotal, err := queries.GetSellerTotalPriceOfAllItems(db, seller.UserId, queries.OnlyHiddenItems)
			require.NoError(t, err)
			require.Equal(t, expectedTotal, actualTotal)
		})

		t.Run("Count only visible items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			expectedTotal := models.MoneyInCents(0)

			// Add visible items
			for i := 0; i < 10; i++ {
				price := models.MoneyInCents((i + 1) * 50)
				setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithFrozen(false), aux.WithHidden(false))
				expectedTotal += price
			}

			// Add hidden items
			for i := 0; i < 10; i++ {
				price := models.MoneyInCents((i + 1) * 150)
				setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithFrozen(false), aux.WithHidden(true))
			}

			actualTotal, err := queries.GetSellerTotalPriceOfAllItems(db, seller.UserId, queries.OnlyVisibleItems)
			require.NoError(t, err)
			require.Equal(t, expectedTotal, actualTotal)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No such user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			nonExistentSellerId := models.Id(1000)
			setup.RequireNoSuchUsers(t, nonExistentSellerId)

			_, err := queries.GetSellerTotalPriceOfAllItems(db, nonExistentSellerId, queries.AllItems)
			require.ErrorIs(t, err, database.ErrNoSuchUser)
		})

		t.Run("Sum of item prices of cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()

			_, err := queries.GetSellerTotalPriceOfAllItems(db, cashier.UserId, queries.AllItems)
			require.ErrorIs(t, err, database.ErrInvalidRole)
		})

		t.Run("Sum of item prices of admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin()

			_, err := queries.GetSellerTotalPriceOfAllItems(db, admin.UserId, queries.AllItems)
			require.ErrorIs(t, err, database.ErrInvalidRole)
		})
	})
}
