//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetItemsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Get only visible items", func(t *testing.T) {
			t.Run("No hidden items", func(t *testing.T) {
				for _, itemCount := range []int{0, 1, 2, 10} {
					testLabel := fmt.Sprintf("Item count: %d", itemCount)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()
						items := setup.Items(seller.UserId, itemCount, aux.WithHidden(false))

						actualItems := []*models.Item{}
						query := queries.NewGetItemsQuery()
						query.WithHidden(false)
						err := query.Execute(db, queries.CollectTo(&actualItems))
						require.NoError(t, err)
						require.Equal(t, itemCount, len(actualItems))

						for i, item := range items {
							require.Equal(t, item, actualItems[i])
						}
					})
				}
			})

			t.Run("With hidden items", func(t *testing.T) {
				for _, itemCount := range []int{0, 1, 2, 10} {
					testLabel := fmt.Sprintf("Item count: %d", itemCount)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()
						setup.Items(seller.UserId, itemCount, aux.WithFrozen(false), aux.WithHidden(true))

						actualItems := []*models.Item{}
						query := queries.NewGetItemsQuery()
						query.WithHidden(false)
						err := query.Execute(db, queries.CollectTo(&actualItems))

						require.NoError(t, err)
						require.Equal(t, 0, len(actualItems))
					})
				}
			})
		})

		t.Run("Get all items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			items = slices.Concat(items, setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(true)))

			actualItems := []*models.Item{}
			query := queries.NewGetItemsQuery()
			err := query.Execute(db, queries.CollectTo(&actualItems))
			require.NoError(t, err)
			require.Equal(t, 20, len(actualItems))

			for i, item := range items {
				require.Equal(t, item, actualItems[i])
			}
		})

		t.Run("Get only hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Items(seller.UserId, 10, aux.WithHidden(false))
			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(true))

			actualItems := []*models.Item{}
			query := queries.NewGetItemsQuery()
			query.WithHidden(true)
			err := query.Execute(db, queries.CollectTo(&actualItems))
			require.NoError(t, err)
			require.Equal(t, 10, len(actualItems))

			for i, item := range items {
				require.Equal(t, item, actualItems[i])
			}
		})

		t.Run("Get items 10-15", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			offset := 10
			limit := 5

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 20, aux.WithHidden(false))

			actualItems := []*models.Item{}
			query := queries.NewGetItemsQuery()
			query.WithLimitAndOffset(uint64(limit), uint64(offset))
			err := query.Execute(db, queries.CollectTo(&actualItems))
			require.NoError(t, err)
			require.Equal(t, limit, len(actualItems))

			for index, actualItem := range actualItems {
				require.Equal(t, items[index+offset], actualItem)
			}
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Failing callback", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			items = slices.Concat(items, setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(true)))

			dummyError := errors.New("test error")
			callback := func(item *models.Item) error {
				return dummyError
			}

			query := queries.NewGetItemsQuery()
			err := query.Execute(db, callback)
			requireDatabaseWrappedError(t, err, dummyError)
		})
	})
}
