//go:build test

package queries

import (
	"bctbackend/algorithms"
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
			t.Run("Count only hidden items", func(t *testing.T) {
				for _, frozenHiddenItemCount := range []int{0, 1, 2} {
					for _, frozenVisibleItemCount := range []int{0, 1, 2} {
						for _, unfrozenHiddenItemCount := range []int{0, 1, 2} {
							for _, unfrozenVisibleItemCount := range []int{0, 1, 2} {
								testLabel := fmt.Sprintf("Seller with %d frozen hidden items, %d frozen visible items, %d unfrozen hidden items, and %d unfrozen visible items", frozenHiddenItemCount, frozenVisibleItemCount, unfrozenHiddenItemCount, unfrozenVisibleItemCount)
								t.Run(testLabel, func(t *testing.T) {
									setup, db := NewDatabaseFixture(WithDefaultCategories)
									defer setup.Close()

									seller := setup.Seller()

									algorithms.Repeat(frozenHiddenItemCount, func() { setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithFrozen(true), aux.WithHidden(true)) })
									algorithms.Repeat(frozenVisibleItemCount, func() { setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithFrozen(true), aux.WithHidden(false)) })
									algorithms.Repeat(unfrozenHiddenItemCount, func() { setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithFrozen(false), aux.WithHidden(true)) })
									algorithms.Repeat(unfrozenVisibleItemCount, func() { setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithFrozen(false), aux.WithHidden(false)) })

									expectedCount := frozenHiddenItemCount

									actual, err := queries.GetSellerFrozenItemCount(db, seller.UserId, queries.OnlyHiddenItems)
									require.NoError(t, err)
									require.Equal(t, expectedCount, actual)
								})
							}
						}
					}
				}
			})

			t.Run("Count only visible items", func(t *testing.T) {
				for _, frozenHiddenItemCount := range []int{0, 1, 2} {
					for _, frozenVisibleItemCount := range []int{0, 1, 2} {
						for _, unfrozenHiddenItemCount := range []int{0, 1, 2} {
							for _, unfrozenVisibleItemCount := range []int{0, 1, 2} {
								testLabel := fmt.Sprintf("Seller with %d frozen hidden items, %d frozen visible items, %d unfrozen hidden items, and %d unfrozen visible items", frozenHiddenItemCount, frozenVisibleItemCount, unfrozenHiddenItemCount, unfrozenVisibleItemCount)
								t.Run(testLabel, func(t *testing.T) {
									setup, db := NewDatabaseFixture(WithDefaultCategories)
									defer setup.Close()

									seller := setup.Seller()

									algorithms.Repeat(frozenHiddenItemCount, func() { setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithFrozen(true), aux.WithHidden(true)) })
									algorithms.Repeat(frozenVisibleItemCount, func() { setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithFrozen(true), aux.WithHidden(false)) })
									algorithms.Repeat(unfrozenHiddenItemCount, func() { setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithFrozen(false), aux.WithHidden(true)) })
									algorithms.Repeat(unfrozenVisibleItemCount, func() { setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithFrozen(false), aux.WithHidden(false)) })

									expectedCount := frozenVisibleItemCount

									actual, err := queries.GetSellerFrozenItemCount(db, seller.UserId, queries.OnlyVisibleItems)
									require.NoError(t, err)
									require.Equal(t, expectedCount, actual)
								})
							}
						}
					}
				}
			})

			t.Run("Count all items", func(t *testing.T) {
				for _, frozenHiddenItemCount := range []int{0, 1, 2} {
					for _, frozenVisibleItemCount := range []int{0, 1, 2} {
						for _, unfrozenHiddenItemCount := range []int{0, 1, 2} {
							for _, unfrozenVisibleItemCount := range []int{0, 1, 2} {
								testLabel := fmt.Sprintf("Seller with %d frozen hidden items, %d frozen visible items, %d unfrozen hidden items, and %d unfrozen visible items", frozenHiddenItemCount, frozenVisibleItemCount, unfrozenHiddenItemCount, unfrozenVisibleItemCount)
								t.Run(testLabel, func(t *testing.T) {
									setup, db := NewDatabaseFixture(WithDefaultCategories)
									defer setup.Close()

									seller := setup.Seller()

									algorithms.Repeat(frozenHiddenItemCount, func() { setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithFrozen(true), aux.WithHidden(true)) })
									algorithms.Repeat(frozenVisibleItemCount, func() { setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithFrozen(true), aux.WithHidden(false)) })
									algorithms.Repeat(unfrozenHiddenItemCount, func() { setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithFrozen(false), aux.WithHidden(true)) })
									algorithms.Repeat(unfrozenVisibleItemCount, func() { setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithFrozen(false), aux.WithHidden(false)) })

									expectedCount := frozenVisibleItemCount + frozenHiddenItemCount

									actual, err := queries.GetSellerFrozenItemCount(db, seller.UserId, queries.AllItems)
									require.NoError(t, err)
									require.Equal(t, expectedCount, actual)
								})
							}
						}
					}
				}
			})
		})
		t.Run("Multiple sellers", func(t *testing.T) {
			for _, frozenItemCount := range []int{0, 1, 2, 10, 100} {
				for _, unfrozenItemCount := range []int{0, 1, 2, 10, 100} {
					for _, hiddenItems := range []bool{false, true} {
						testLabel := fmt.Sprintf("Seller with %d frozen items and %d unfrozen items; hidden=%v", frozenItemCount, unfrozenItemCount, hiddenItems)
						t.Run(testLabel, func(t *testing.T) {
							setup, db := NewDatabaseFixture(WithDefaultCategories)
							defer setup.Close()

							seller := setup.Seller()
							otherSeller := setup.Seller()

							for i := 0; i < frozenItemCount; i++ {
								setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithFrozen(true), aux.WithHidden(hiddenItems))
								setup.Item(otherSeller.UserId, aux.WithDummyData(int(i)), aux.WithFrozen(true), aux.WithHidden(hiddenItems))
							}
							for i := 0; i < unfrozenItemCount; i++ {
								setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithFrozen(false), aux.WithHidden(hiddenItems))
								setup.Item(otherSeller.UserId, aux.WithDummyData(int(i)), aux.WithFrozen(false), aux.WithHidden(hiddenItems))
							}

							actual, err := queries.GetSellerFrozenItemCount(db, seller.UserId, queries.AllItems)
							require.NoError(t, err)
							require.Equal(t, frozenItemCount, actual)
						})
					}
				}
			}
		})

		t.Run("Failure", func(t *testing.T) {
			t.Run("No such user", func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				nonExistentSellerId := models.Id(1000)
				setup.RequireNoSuchUser(t, nonExistentSellerId)

				_, err := queries.GetSellerFrozenItemCount(db, nonExistentSellerId, queries.AllItems)
				{
					var noSuchUserError *queries.NoSuchUserError
					require.ErrorAs(t, err, &noSuchUserError)
				}
			})

			t.Run("Count frozen items of cashier", func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				cashier := setup.Cashier()

				_, err := queries.GetSellerFrozenItemCount(db, cashier.UserId, queries.AllItems)
				{
					var invalidRoleError *queries.InvalidRoleError
					require.ErrorAs(t, err, &invalidRoleError)
				}
			})

			t.Run("Count frozen items of admin", func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				admin := setup.Admin()

				_, err := queries.GetSellerFrozenItemCount(db, admin.UserId, queries.AllItems)
				{
					var invalidRoleError *queries.InvalidRoleError
					require.ErrorAs(t, err, &invalidRoleError)
				}
			})
		})
	})
}
