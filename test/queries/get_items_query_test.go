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

func TestListItemsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Get only visible items", func(t *testing.T) {
			t.Run("No hidden items in database", func(t *testing.T) {
				for _, itemCount := range []int{0, 1, 2, 10} {
					testLabel := fmt.Sprintf("Item count: %d", itemCount)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()
						items := setup.Items(seller.UserID, itemCount, aux.WithHidden(false))

						query := queries.ListItems()
						query.WithHidden(false)
						actualItems := []*models.Item{}
						err := query.Execute(db, queries.CollectTo(&actualItems))
						require.NoError(t, err)
						require.Equal(t, itemCount, len(actualItems))

						for i, item := range items {
							require.Equal(t, item, actualItems[i])
						}
					})
				}
			})

			t.Run("Only hidden items in database", func(t *testing.T) {
				for _, itemCount := range []int{0, 1, 2, 10} {
					testLabel := fmt.Sprintf("Item count: %d", itemCount)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()
						setup.Items(seller.UserID, itemCount, aux.WithFrozen(false), aux.WithHidden(true))

						query := queries.ListItems()
						query.WithHidden(false)

						actualItems := []*models.Item{}
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
			items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
			items = slices.Concat(items, setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(true)))

			actualItems := []*models.Item{}
			query := queries.ListItems()
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
			setup.Items(seller.UserID, 10, aux.WithHidden(false))
			items := setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(true))

			actualItems := []*models.Item{}
			query := queries.ListItems()
			query.WithHidden(true)
			err := query.Execute(db, queries.CollectTo(&actualItems))
			require.NoError(t, err)
			require.Equal(t, 10, len(actualItems))

			for i, item := range items {
				require.Equal(t, item, actualItems[i])
			}
		})

		// t.Run("Get items 10-15", func(t *testing.T) {
		// 	setup, db := NewDatabaseFixture(WithDefaultCategories)
		// 	defer setup.Close()

		// 	offset := uint64(10)
		// 	limit := uint64(5)

		// 	seller := setup.Seller()
		// 	items := setup.Items(seller.UserID, 20, aux.WithHidden(false))

		// 	actualItems := []*models.Item{}
		// 	query := queries.NewGetItemsQuery()
		// 	query.WithRowRange(&queries.RowRange{Limit: &limit, Offset: &offset})
		// 	err := query.Execute(db, queries.CollectTo(&actualItems))
		// 	require.NoError(t, err)
		// 	require.Equal(t, int(limit), len(actualItems))

		// 	for index, actualItem := range actualItems {
		// 		require.Equal(t, items[index+int(offset)], actualItem)
		// 	}
		// })

		t.Run("Get items of specific category", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			setup.Category(1, "a")
			setup.Category(2, "b")
			setup.Category(3, "c")
			items1 := setup.Items(seller.UserID, 1, aux.WithHidden(false), aux.WithItemCategory(1))
			items2 := setup.Items(seller.UserID, 2, aux.WithHidden(false), aux.WithItemCategory(2))
			items3 := setup.Items(seller.UserID, 3, aux.WithHidden(false), aux.WithItemCategory(3))

			{
				actualItems := []*models.Item{}
				query := queries.ListItems()
				query.WithCategory(1)
				err := query.Execute(db, queries.CollectTo(&actualItems))
				require.NoError(t, err)
				require.Len(t, actualItems, 1)
				require.ElementsMatch(t, items1, actualItems)
			}

			{
				actualItems := []*models.Item{}
				query := queries.ListItems()
				query.WithCategory(2)
				err := query.Execute(db, queries.CollectTo(&actualItems))
				require.NoError(t, err)
				require.Len(t, actualItems, 2)
				require.ElementsMatch(t, items2, actualItems)
			}

			{
				actualItems := []*models.Item{}
				query := queries.ListItems()
				query.WithCategory(3)
				err := query.Execute(db, queries.CollectTo(&actualItems))
				require.NoError(t, err)
				require.Len(t, actualItems, 3)
				require.ElementsMatch(t, items3, actualItems)
			}
		})

		t.Run("Get items with description containing a", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Item(seller.UserID, aux.WithDescription("foo"), aux.WithHidden(false))
			item2 := setup.Item(seller.UserID, aux.WithDescription("bar"), aux.WithHidden(false))
			item3 := setup.Item(seller.UserID, aux.WithDescription("baz"), aux.WithHidden(false))
			setup.Item(seller.UserID, aux.WithDescription("qux"), aux.WithHidden(false))

			actualItems := []*models.Item{}
			query := queries.ListItems()
			query.WithDescriptionPattern("%a%")
			require.NoError(t, query.Execute(db, queries.CollectTo(&actualItems)))
			require.Equal(t, 2, len(actualItems))
			require.Equal(t, item2.ItemID, actualItems[0].ItemID)
			require.Equal(t, item3.ItemID, actualItems[1].ItemID)
		})

		// t.Run("Get items with description containing a space", func(t *testing.T) {
		// 	setup, db := NewDatabaseFixture(WithDefaultCategories)
		// 	defer setup.Close()

		// 	seller := setup.Seller()
		// 	setup.Item(seller.UserID, aux.WithDescription("foo"), aux.WithHidden(false))
		// 	item2 := setup.Item(seller.UserID, aux.WithDescription("b ar"), aux.WithHidden(false))
		// 	item3 := setup.Item(seller.UserID, aux.WithDescription("ba z"), aux.WithHidden(false))
		// 	setup.Item(seller.UserID, aux.WithDescription("qux"), aux.WithHidden(false))
		// 	item5 := setup.Item(seller.UserID, aux.WithDescription("q u x"), aux.WithHidden(false))

		// 	actualItems := []*models.Item{}
		// 	query := queries.NewGetItemsQuery()
		// 	query.WithDescriptionPattern("% %")
		// 	require.NoError(t, query.Execute(db, queries.CollectTo(&actualItems)))
		// 	require.Equal(t, 3, len(actualItems))
		// 	require.Equal(t, item2.ItemID, actualItems[0].ItemID)
		// 	require.Equal(t, item3.ItemID, actualItems[1].ItemID)
		// 	require.Equal(t, item5.ItemID, actualItems[2].ItemID)
		// })

		// t.Run("Get items with description containing an ampersand", func(t *testing.T) {
		// 	setup, db := NewDatabaseFixture(WithDefaultCategories)
		// 	defer setup.Close()

		// 	seller := setup.Seller()
		// 	setup.Item(seller.UserID, aux.WithDescription("foo"), aux.WithHidden(false))
		// 	item2 := setup.Item(seller.UserID, aux.WithDescription("b&ar"), aux.WithHidden(false))
		// 	item3 := setup.Item(seller.UserID, aux.WithDescription("ba&z"), aux.WithHidden(false))
		// 	setup.Item(seller.UserID, aux.WithDescription("qux"), aux.WithHidden(false))
		// 	item5 := setup.Item(seller.UserID, aux.WithDescription("q u&x"), aux.WithHidden(false))

		// 	actualItems := []*models.Item{}
		// 	query := queries.NewGetItemsQuery()
		// 	query.WithDescriptionPattern("%&%")
		// 	require.NoError(t, query.Execute(db, queries.CollectTo(&actualItems)))
		// 	require.Equal(t, 3, len(actualItems))
		// 	require.Equal(t, item2.ItemID, actualItems[0].ItemID)
		// 	require.Equal(t, item3.ItemID, actualItems[1].ItemID)
		// 	require.Equal(t, item5.ItemID, actualItems[2].ItemID)
		// })

		// t.Run("Get small items", func(t *testing.T) {
		// 	setup, db := NewDatabaseFixture(WithDefaultCategories)
		// 	defer setup.Close()

		// 	seller := setup.Seller()
		// 	item1 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	item2 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
		// 	item3 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
		// 	item4 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	item5 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))

		// 	actualItems := []*models.Item{}
		// 	query := queries.NewGetItemsQuery()
		// 	query.WithLarge(false)
		// 	err := query.Execute(db, queries.CollectTo(&actualItems))
		// 	require.NoError(t, err)
		// 	require.Len(t, actualItems, 5)
		// 	require.Equal(t, item1.ItemID, actualItems[0].ItemID)
		// 	require.Equal(t, item2.ItemID, actualItems[1].ItemID)
		// 	require.Equal(t, item3.ItemID, actualItems[2].ItemID)
		// 	require.Equal(t, item4.ItemID, actualItems[3].ItemID)
		// 	require.Equal(t, item5.ItemID, actualItems[4].ItemID)
		// })

		// t.Run("Get large items", func(t *testing.T) {
		// 	setup, db := NewDatabaseFixture(WithDefaultCategories)
		// 	defer setup.Close()

		// 	seller := setup.Seller()
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	item1 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	item2 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
		// 	item3 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	item4 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))

		// 	actualItems := []*models.Item{}
		// 	query := queries.NewGetItemsQuery()
		// 	query.WithLarge(true)
		// 	err := query.Execute(db, queries.CollectTo(&actualItems))
		// 	require.NoError(t, err)
		// 	require.Len(t, actualItems, 4)
		// 	require.Equal(t, item1.ItemID, actualItems[0].ItemID)
		// 	require.Equal(t, item2.ItemID, actualItems[1].ItemID)
		// 	require.Equal(t, item3.ItemID, actualItems[2].ItemID)
		// 	require.Equal(t, item4.ItemID, actualItems[3].ItemID)
		// })

		// t.Run("Get item selection", func(t *testing.T) {
		// 	setup, db := NewDatabaseFixture(WithDefaultCategories)
		// 	defer setup.Close()

		// 	seller := setup.Seller()
		// 	item1 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	item2 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
		// 	item3 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
		// 	setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))

		// 	actualItems := []*models.Item{}
		// 	query := queries.NewGetItemsQuery()
		// 	query.WithItemIDs([]models.ID{item1.ItemID, item2.ItemID, item3.ItemID})
		// 	err := query.Execute(db, queries.CollectTo(&actualItems))
		// 	require.NoError(t, err)
		// 	require.Len(t, actualItems, 3)
		// 	require.Equal(t, item1.ItemID, actualItems[0].ItemID)
		// 	require.Equal(t, item2.ItemID, actualItems[1].ItemID)
		// 	require.Equal(t, item3.ItemID, actualItems[2].ItemID)
		// })
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Failing callback", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
			items = slices.Concat(items, setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(true)))

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
						items := setup.Items(seller.UserID, itemCount, aux.WithHidden(false))

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
						setup.Items(seller.UserID, itemCount, aux.WithFrozen(false), aux.WithHidden(true))

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
			items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
			items = slices.Concat(items, setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(true)))

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
			setup.Items(seller.UserID, 10, aux.WithHidden(false))
			items := setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(true))

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

			offset := uint64(10)
			limit := uint64(5)

			seller := setup.Seller()
			items := setup.Items(seller.UserID, 20, aux.WithHidden(false))

			actualItems := []*models.Item{}
			query := queries.NewGetItemsQuery()
			query.WithRowRange(&queries.RowRange{Limit: &limit, Offset: &offset})
			err := query.Execute(db, queries.CollectTo(&actualItems))
			require.NoError(t, err)
			require.Equal(t, int(limit), len(actualItems))

			for index, actualItem := range actualItems {
				require.Equal(t, items[index+int(offset)], actualItem)
			}
		})

		t.Run("Get items of specific category", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			setup.Category(1, "a")
			setup.Category(2, "b")
			setup.Category(3, "c")
			items1 := setup.Items(seller.UserID, 1, aux.WithHidden(false), aux.WithItemCategory(1))
			items2 := setup.Items(seller.UserID, 2, aux.WithHidden(false), aux.WithItemCategory(2))
			items3 := setup.Items(seller.UserID, 3, aux.WithHidden(false), aux.WithItemCategory(3))

			{
				actualItems := []*models.Item{}
				query := queries.NewGetItemsQuery()
				query.WithCategory(1)
				err := query.Execute(db, queries.CollectTo(&actualItems))
				require.NoError(t, err)
				require.Len(t, actualItems, 1)
				require.ElementsMatch(t, items1, actualItems)
			}

			{
				actualItems := []*models.Item{}
				query := queries.NewGetItemsQuery()
				query.WithCategory(2)
				err := query.Execute(db, queries.CollectTo(&actualItems))
				require.NoError(t, err)
				require.Len(t, actualItems, 2)
				require.ElementsMatch(t, items2, actualItems)
			}

			{
				actualItems := []*models.Item{}
				query := queries.NewGetItemsQuery()
				query.WithCategory(3)
				err := query.Execute(db, queries.CollectTo(&actualItems))
				require.NoError(t, err)
				require.Len(t, actualItems, 3)
				require.ElementsMatch(t, items3, actualItems)
			}
		})

		t.Run("Get items with description containing a", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Item(seller.UserID, aux.WithDescription("foo"), aux.WithHidden(false))
			item2 := setup.Item(seller.UserID, aux.WithDescription("bar"), aux.WithHidden(false))
			item3 := setup.Item(seller.UserID, aux.WithDescription("baz"), aux.WithHidden(false))
			setup.Item(seller.UserID, aux.WithDescription("qux"), aux.WithHidden(false))

			actualItems := []*models.Item{}
			query := queries.NewGetItemsQuery()
			query.WithDescriptionPattern("%a%")
			require.NoError(t, query.Execute(db, queries.CollectTo(&actualItems)))
			require.Equal(t, 2, len(actualItems))
			require.Equal(t, item2.ItemID, actualItems[0].ItemID)
			require.Equal(t, item3.ItemID, actualItems[1].ItemID)
		})

		t.Run("Get items with description containing a space", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Item(seller.UserID, aux.WithDescription("foo"), aux.WithHidden(false))
			item2 := setup.Item(seller.UserID, aux.WithDescription("b ar"), aux.WithHidden(false))
			item3 := setup.Item(seller.UserID, aux.WithDescription("ba z"), aux.WithHidden(false))
			setup.Item(seller.UserID, aux.WithDescription("qux"), aux.WithHidden(false))
			item5 := setup.Item(seller.UserID, aux.WithDescription("q u x"), aux.WithHidden(false))

			actualItems := []*models.Item{}
			query := queries.NewGetItemsQuery()
			query.WithDescriptionPattern("% %")
			require.NoError(t, query.Execute(db, queries.CollectTo(&actualItems)))
			require.Equal(t, 3, len(actualItems))
			require.Equal(t, item2.ItemID, actualItems[0].ItemID)
			require.Equal(t, item3.ItemID, actualItems[1].ItemID)
			require.Equal(t, item5.ItemID, actualItems[2].ItemID)
		})

		t.Run("Get items with description containing an ampersand", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Item(seller.UserID, aux.WithDescription("foo"), aux.WithHidden(false))
			item2 := setup.Item(seller.UserID, aux.WithDescription("b&ar"), aux.WithHidden(false))
			item3 := setup.Item(seller.UserID, aux.WithDescription("ba&z"), aux.WithHidden(false))
			setup.Item(seller.UserID, aux.WithDescription("qux"), aux.WithHidden(false))
			item5 := setup.Item(seller.UserID, aux.WithDescription("q u&x"), aux.WithHidden(false))

			actualItems := []*models.Item{}
			query := queries.NewGetItemsQuery()
			query.WithDescriptionPattern("%&%")
			require.NoError(t, query.Execute(db, queries.CollectTo(&actualItems)))
			require.Equal(t, 3, len(actualItems))
			require.Equal(t, item2.ItemID, actualItems[0].ItemID)
			require.Equal(t, item3.ItemID, actualItems[1].ItemID)
			require.Equal(t, item5.ItemID, actualItems[2].ItemID)
		})

		t.Run("Get small items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			item1 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			item2 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
			item3 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
			item4 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			item5 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))

			actualItems := []*models.Item{}
			query := queries.NewGetItemsQuery()
			query.WithLarge(false)
			err := query.Execute(db, queries.CollectTo(&actualItems))
			require.NoError(t, err)
			require.Len(t, actualItems, 5)
			require.Equal(t, item1.ItemID, actualItems[0].ItemID)
			require.Equal(t, item2.ItemID, actualItems[1].ItemID)
			require.Equal(t, item3.ItemID, actualItems[2].ItemID)
			require.Equal(t, item4.ItemID, actualItems[3].ItemID)
			require.Equal(t, item5.ItemID, actualItems[4].ItemID)
		})

		t.Run("Get large items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			item1 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			item2 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
			item3 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			item4 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))

			actualItems := []*models.Item{}
			query := queries.NewGetItemsQuery()
			query.WithLarge(true)
			err := query.Execute(db, queries.CollectTo(&actualItems))
			require.NoError(t, err)
			require.Len(t, actualItems, 4)
			require.Equal(t, item1.ItemID, actualItems[0].ItemID)
			require.Equal(t, item2.ItemID, actualItems[1].ItemID)
			require.Equal(t, item3.ItemID, actualItems[2].ItemID)
			require.Equal(t, item4.ItemID, actualItems[3].ItemID)
		})

		t.Run("Get item selection", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			item1 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			item2 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))
			item3 := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(true))
			setup.Item(seller.UserID, aux.WithHidden(false), aux.WithLarge(false))

			actualItems := []*models.Item{}
			query := queries.NewGetItemsQuery()
			query.WithItemIDs([]models.ID{item1.ItemID, item2.ItemID, item3.ItemID})
			err := query.Execute(db, queries.CollectTo(&actualItems))
			require.NoError(t, err)
			require.Len(t, actualItems, 3)
			require.Equal(t, item1.ItemID, actualItems[0].ItemID)
			require.Equal(t, item2.ItemID, actualItems[1].ItemID)
			require.Equal(t, item3.ItemID, actualItems[2].ItemID)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Failing callback", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
			items = slices.Concat(items, setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(true)))

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
