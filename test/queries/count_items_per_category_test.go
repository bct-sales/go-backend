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
)

func TestGetCategoryCounts(t *testing.T) {
	defaultCategoryNameTable := aux.DefaultCategoryNameTable()

	t.Run("Success", func(t *testing.T) {
		t.Run("Without hidden items", func(t *testing.T) {
			countTables := []map[models.ID]int{
				{},
				{
					aux.CategoryID_Clothing50_56: 1,
				},
				{
					aux.CategoryID_Clothing50_56: 2,
				},
				{
					aux.CategoryID_Clothing50_56: 2,
					aux.CategoryID_Toys:          3,
				},
				{
					aux.CategoryID_Clothing50_56:      1,
					aux.CategoryID_Clothing56_62:      2,
					aux.CategoryID_Clothing68_80:      3,
					aux.CategoryID_Clothing86_92:      4,
					aux.CategoryID_Clothing92_98:      5,
					aux.CategoryID_Clothing104_116:    6,
					aux.CategoryID_Clothing122_128:    7,
					aux.CategoryID_Clothing128_140:    8,
					aux.CategoryID_Clothing140_152:    9,
					aux.CategoryID_Shoes:              10,
					aux.CategoryID_Toys:               11,
					aux.CategoryID_BabyChildEquipment: 12,
				},
			}

			for _, expectedCounts := range countTables {
				testLabel := fmt.Sprintf("Count table: %v", expectedCounts)

				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()

					for categoryId, count := range expectedCounts {
						for i := 0; i < count; i++ {
							setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithItemCategory(categoryId), aux.WithHidden(false))
						}
					}

					actualCounts, err := queries.CountItemsPerCategory(db, queries.AllItems)
					require.NoError(t, err)
					require.Equal(t, len(defaultCategoryNameTable), len(actualCounts))

					for categoryId, _ := range defaultCategoryNameTable {
						actualCount, ok := actualCounts[categoryId]
						require.True(t, ok, "Category ID %d not found in actual counts", categoryId)
						expectedCount := expectedCounts[categoryId]

						require.Equal(t, expectedCount, actualCount)
					}
				})
			}
		})

		t.Run("With hidden items", func(t *testing.T) {
			t.Run("Only visible items", func(t *testing.T) {
				countTables := []map[models.ID]int{
					{},
					{
						aux.CategoryID_Clothing50_56: 1,
					},
					{
						aux.CategoryID_Clothing50_56: 2,
					},
					{
						aux.CategoryID_Clothing50_56: 2,
						aux.CategoryID_Toys:          3,
					},
					{
						aux.CategoryID_Clothing50_56:      1,
						aux.CategoryID_Clothing56_62:      2,
						aux.CategoryID_Clothing68_80:      3,
						aux.CategoryID_Clothing86_92:      4,
						aux.CategoryID_Clothing92_98:      5,
						aux.CategoryID_Clothing104_116:    6,
						aux.CategoryID_Clothing122_128:    7,
						aux.CategoryID_Clothing128_140:    8,
						aux.CategoryID_Clothing140_152:    9,
						aux.CategoryID_Shoes:              10,
						aux.CategoryID_Toys:               11,
						aux.CategoryID_BabyChildEquipment: 12,
					},
				}

				for _, expectedCounts := range countTables {
					testLabel := fmt.Sprintf("Count table %v", expectedCounts)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()

						for categoryId, count := range expectedCounts {
							for i := 0; i < count; i++ {
								setup.Item(seller.UserID, aux.WithDummyData(i), aux.WithItemCategory(categoryId), aux.WithFrozen(false), aux.WithHidden(false))
								setup.Item(seller.UserID, aux.WithDummyData(2*i), aux.WithItemCategory(categoryId), aux.WithFrozen(false), aux.WithHidden(true))
								setup.Item(seller.UserID, aux.WithDummyData(3*i), aux.WithItemCategory(categoryId), aux.WithFrozen(false), aux.WithHidden(true))
							}
						}

						actualCounts, err := queries.CountItemsPerCategory(db, queries.OnlyVisibleItems)
						require.NoError(t, err)
						require.Equal(t, len(defaultCategoryNameTable), len(actualCounts))

						for categoryId, _ := range defaultCategoryNameTable {
							actualCount, ok := actualCounts[categoryId]
							require.True(t, ok, "Category ID %d not found in actual counts", categoryId)
							expectedCount := expectedCounts[categoryId]

							require.Equal(t, expectedCount, actualCount, "Wrong count for category %d", categoryId)
						}
					})
				}
			})

			t.Run("Only hidden items", func(t *testing.T) {
				countTables := []map[models.ID]int{
					{},
					{
						aux.CategoryID_Clothing50_56: 1,
					},
					{
						aux.CategoryID_Clothing50_56: 2,
					},
					{
						aux.CategoryID_Clothing50_56: 2,
						aux.CategoryID_Toys:          3,
					},
					{
						aux.CategoryID_Clothing50_56:      1,
						aux.CategoryID_Clothing56_62:      2,
						aux.CategoryID_Clothing68_80:      3,
						aux.CategoryID_Clothing86_92:      4,
						aux.CategoryID_Clothing92_98:      5,
						aux.CategoryID_Clothing104_116:    6,
						aux.CategoryID_Clothing122_128:    7,
						aux.CategoryID_Clothing128_140:    8,
						aux.CategoryID_Clothing140_152:    9,
						aux.CategoryID_Shoes:              10,
						aux.CategoryID_Toys:               11,
						aux.CategoryID_BabyChildEquipment: 12,
					},
				}

				for _, expectedCounts := range countTables {
					testLabel := fmt.Sprintf("Count table %v", expectedCounts)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()

						for categoryId, count := range expectedCounts {
							for i := 0; i < count; i++ {
								setup.Item(seller.UserID, aux.WithDummyData(i), aux.WithItemCategory(categoryId), aux.WithFrozen(false), aux.WithHidden(false))
								setup.Item(seller.UserID, aux.WithDummyData(2*i), aux.WithItemCategory(categoryId), aux.WithFrozen(false), aux.WithHidden(true))
								setup.Item(seller.UserID, aux.WithDummyData(3*i), aux.WithItemCategory(categoryId), aux.WithFrozen(false), aux.WithHidden(true))
							}
						}

						actualCounts, err := queries.CountItemsPerCategory(db, queries.OnlyHiddenItems)
						require.NoError(t, err)
						require.Equal(t, len(defaultCategoryNameTable), len(actualCounts))

						for categoryId, _ := range defaultCategoryNameTable {
							actualCount, ok := actualCounts[categoryId]
							require.True(t, ok, "Category ID %d not found in actual counts", categoryId)
							expectedCount := expectedCounts[categoryId] * 2

							require.Equal(t, expectedCount, actualCount, "Wrong count for category %d", categoryId)
						}
					})
				}
			})

			t.Run("All items", func(t *testing.T) {
				countTables := []map[models.ID]int{
					{},
					{
						aux.CategoryID_Clothing50_56: 1,
					},
					{
						aux.CategoryID_Clothing50_56: 2,
					},
					{
						aux.CategoryID_Clothing50_56: 2,
						aux.CategoryID_Toys:          3,
					},
					{
						aux.CategoryID_Clothing50_56:      1,
						aux.CategoryID_Clothing56_62:      2,
						aux.CategoryID_Clothing68_80:      3,
						aux.CategoryID_Clothing86_92:      4,
						aux.CategoryID_Clothing92_98:      5,
						aux.CategoryID_Clothing104_116:    6,
						aux.CategoryID_Clothing122_128:    7,
						aux.CategoryID_Clothing128_140:    8,
						aux.CategoryID_Clothing140_152:    9,
						aux.CategoryID_Shoes:              10,
						aux.CategoryID_Toys:               11,
						aux.CategoryID_BabyChildEquipment: 12,
					},
				}

				for _, expectedCounts := range countTables {
					testLabel := fmt.Sprintf("Count table %v", expectedCounts)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()

						for categoryId, count := range expectedCounts {
							for i := 0; i < count; i++ {
								setup.Item(seller.UserID, aux.WithDummyData(i), aux.WithItemCategory(categoryId), aux.WithFrozen(false), aux.WithHidden(false))
								setup.Item(seller.UserID, aux.WithDummyData(2*i), aux.WithItemCategory(categoryId), aux.WithFrozen(false), aux.WithHidden(true))
								setup.Item(seller.UserID, aux.WithDummyData(3*i), aux.WithItemCategory(categoryId), aux.WithFrozen(false), aux.WithHidden(true))
							}
						}

						actualCounts, err := queries.CountItemsPerCategory(db, queries.AllItems)
						require.NoError(t, err)
						require.Equal(t, len(defaultCategoryNameTable), len(actualCounts))

						for categoryId, _ := range defaultCategoryNameTable {
							actualCount, ok := actualCounts[categoryId]
							require.True(t, ok, "Category ID %d not found in actual counts", categoryId)
							expectedCount := expectedCounts[categoryId] * 3

							require.Equal(t, expectedCount, actualCount)
						}
					})
				}
			})
		})
	})
}
