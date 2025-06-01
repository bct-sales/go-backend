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

func TestGetSellerFrozenItemCount(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Single seller", func(t *testing.T) {
			t.Parallel()

			for _, frozenVisibleItemCount := range []int{0, 1, 2} {
				for _, unfrozenHiddenItemCount := range []int{0, 1, 2} {
					for _, unfrozenVisibleItemCount := range []int{0, 1, 2} {
						testLabel := fmt.Sprintf("Seller with %d frozen visible items, %d unfrozen hidden items, and %d unfrozen visible items", frozenVisibleItemCount, unfrozenHiddenItemCount, unfrozenVisibleItemCount)
						t.Run(testLabel, func(t *testing.T) {
							t.Parallel()

							setup, db := NewDatabaseFixture(WithDefaultCategories)
							defer setup.Close()

							seller := setup.Seller()

							setup.Items(seller.UserId, frozenVisibleItemCount, aux.WithDummyData(0), aux.WithFrozen(true), aux.WithHidden(false))
							setup.Items(seller.UserId, unfrozenHiddenItemCount, aux.WithDummyData(0), aux.WithFrozen(false), aux.WithHidden(true))
							setup.Items(seller.UserId, unfrozenVisibleItemCount, aux.WithDummyData(0), aux.WithFrozen(false), aux.WithHidden(false))

							expectedCount := frozenVisibleItemCount

							actual, err := queries.GetSellerFrozenItemCount(db, seller.UserId)
							require.NoError(t, err)
							require.Equal(t, expectedCount, actual)
						})
					}
				}
			}
		})
		t.Run("Multiple sellers", func(t *testing.T) {
			for _, frozenItemCount := range []int{0, 1, 2, 10, 100} {
				for _, unfrozenItemCount := range []int{0, 1, 2, 10, 100} {
					testLabel := fmt.Sprintf("Seller with %d frozen items and %d unfrozen items", frozenItemCount, unfrozenItemCount)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()
						otherSeller := setup.Seller()

						for i := 0; i < frozenItemCount; i++ {
							setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithFrozen(true), aux.WithHidden(false))
							setup.Item(otherSeller.UserId, aux.WithDummyData(int(i)), aux.WithFrozen(true), aux.WithHidden(false))
						}
						for i := 0; i < unfrozenItemCount; i++ {
							setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithFrozen(false), aux.WithHidden(false))
							setup.Item(otherSeller.UserId, aux.WithDummyData(int(i)), aux.WithFrozen(false), aux.WithHidden(false))
						}

						actual, err := queries.GetSellerFrozenItemCount(db, seller.UserId)
						require.NoError(t, err)
						require.Equal(t, frozenItemCount, actual)
					})
				}
			}
		})

		t.Run("Failure", func(t *testing.T) {
			t.Run("No such user", func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				nonExistentSellerId := models.Id(1000)
				setup.RequireNoSuchUsers(t, nonExistentSellerId)

				_, err := queries.GetSellerFrozenItemCount(db, nonExistentSellerId)
				require.ErrorIs(t, err, database.ErrNoSuchUser)
			})

			t.Run("Count frozen items of cashier", func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				cashier := setup.Cashier()

				_, err := queries.GetSellerFrozenItemCount(db, cashier.UserId)
				require.ErrorIs(t, err, database.ErrWrongRole)
			})

			t.Run("Count frozen items of admin", func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				admin := setup.Admin()

				_, err := queries.GetSellerFrozenItemCount(db, admin.UserId)
				require.ErrorIs(t, err, database.ErrWrongRole)
			})
		})
	})
}
