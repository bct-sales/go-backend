//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSellerTotalValueOfAllTimes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Single seller", func(t *testing.T) {
			for _, itemCount := range []int64{0, 1, 2, 10, 100} {
				testLabel := fmt.Sprintf("Seller with %d items", itemCount)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture(t, WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller(t)

					expectedTotal := models.MoneyInCents(0)
					for i := int64(0); i < itemCount; i++ {
						price := models.MoneyInCents((i + 1) * 50)
						setup.Item(t, seller.UserID, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithHidden(false))
						expectedTotal += price
					}

					actualTotal, err := queries.GetSellerTotalValueOfAllItems(t.Context(), db, seller.UserID, queries.AllItems)
					require.NoError(t, err)
					require.Equal(t, expectedTotal, actualTotal)
				})
			}
		})

		t.Run("Multiple seller", func(t *testing.T) {
			for _, itemCount := range []int64{0, 1, 2, 10, 100} {
				testLabel := fmt.Sprintf("Seller with %d items", itemCount)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture(t, WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller(t)
					otherSeller := setup.Seller(t)

					expectedTotal := models.MoneyInCents(0)
					for i := int64(0); i < itemCount; i++ {
						price := models.MoneyInCents((i + 1) * 50)
						setup.Item(t, seller.UserID, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithHidden(false))
						setup.Item(t, otherSeller.UserID, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithHidden(false))
						expectedTotal += price
					}

					actualTotal, err := queries.GetSellerTotalValueOfAllItems(t.Context(), db, seller.UserID, queries.AllItems)
					require.NoError(t, err)
					require.Equal(t, expectedTotal, actualTotal)
				})
			}
		})

		t.Run("Count all items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)

			expectedTotal := models.MoneyInCents(0)

			// Add visible items
			for i := 0; i < 10; i++ {
				price := models.MoneyInCents((i + 1) * 50)
				setup.Item(t, seller.UserID, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithFrozen(false), aux.WithHidden(false))
				expectedTotal += price
			}

			// Add hidden items
			for i := 0; i < 10; i++ {
				price := models.MoneyInCents((i + 1) * 150)
				setup.Item(t, seller.UserID, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithFrozen(false), aux.WithHidden(true))
				expectedTotal += price
			}

			actualTotal, err := queries.GetSellerTotalValueOfAllItems(t.Context(), db, seller.UserID, queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, expectedTotal, actualTotal)
		})

		t.Run("Count only hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)

			expectedTotal := models.MoneyInCents(0)

			// Add visible items
			for i := 0; i < 10; i++ {
				price := models.MoneyInCents((i + 1) * 50)
				setup.Item(t, seller.UserID, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithFrozen(false), aux.WithHidden(false))
			}

			// Add hidden items
			for i := 0; i < 10; i++ {
				price := models.MoneyInCents((i + 1) * 150)
				setup.Item(t, seller.UserID, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithFrozen(false), aux.WithHidden(true))
				expectedTotal += price
			}

			actualTotal, err := queries.GetSellerTotalValueOfAllItems(t.Context(), db, seller.UserID, queries.OnlyHiddenItems)
			require.NoError(t, err)
			require.Equal(t, expectedTotal, actualTotal)
		})

		t.Run("Count only visible items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)

			expectedTotal := models.MoneyInCents(0)

			// Add visible items
			for i := 0; i < 10; i++ {
				price := models.MoneyInCents((i + 1) * 50)
				setup.Item(t, seller.UserID, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithFrozen(false), aux.WithHidden(false))
				expectedTotal += price
			}

			// Add hidden items
			for i := 0; i < 10; i++ {
				price := models.MoneyInCents((i + 1) * 150)
				setup.Item(t, seller.UserID, aux.WithDummyData(int(i)), aux.WithPriceInCents(price), aux.WithFrozen(false), aux.WithHidden(true))
			}

			actualTotal, err := queries.GetSellerTotalValueOfAllItems(t.Context(), db, seller.UserID, queries.OnlyVisibleItems)
			require.NoError(t, err)
			require.Equal(t, expectedTotal, actualTotal)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No such user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			nonExistentSellerID := setup.GenerateNonexistentUserID(t)

			_, err := queries.GetSellerTotalValueOfAllItems(t.Context(), db, nonExistentSellerID, queries.AllItems)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})

		t.Run("Sum of item prices of cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier(t)

			_, err := queries.GetSellerTotalValueOfAllItems(t.Context(), db, cashier.UserID, queries.AllItems)
			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})

		t.Run("Sum of item prices of admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin(t)

			_, err := queries.GetSellerTotalValueOfAllItems(t.Context(), db, admin.UserID, queries.AllItems)
			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})
	})
}
