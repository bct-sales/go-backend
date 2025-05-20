//go:build test

package queries

import (
	"bctbackend/database/queries"
	"bctbackend/defs"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetCategoryCounts(t *testing.T) {
	defaultCategoryTable := DefaultCategoryTable()

	t.Run("Success", func(t *testing.T) {
		t.Run("Without hidden items", func(t *testing.T) {
			countTables := []map[defs.Id]int64{
				{},
				{
					CategoryId_Clothing50_56: 1,
				},
				{
					CategoryId_Clothing50_56: 2,
				},
				{
					CategoryId_Clothing50_56: 2,
					CategoryId_Toys:          3,
				},
				{
					CategoryId_Clothing50_56:      1,
					CategoryId_Clothing56_62:      2,
					CategoryId_Clothing68_80:      3,
					CategoryId_Clothing86_92:      4,
					CategoryId_Clothing92_98:      5,
					CategoryId_Clothing104_116:    6,
					CategoryId_Clothing122_128:    7,
					CategoryId_Clothing128_140:    8,
					CategoryId_Clothing140_152:    9,
					CategoryId_Shoes:              10,
					CategoryId_Toys:               11,
					CategoryId_BabyChildEquipment: 12,
				},
			}

			for _, expectedCountTable := range countTables {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()

				for categoryId, count := range expectedCountTable {
					for i := int64(0); i < count; i++ {
						setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithItemCategory(categoryId), aux.WithHidden(false))
					}
				}

				actualCounts, err := queries.GetCategoryCounts(db, true)
				require.NoError(t, err)
				require.Equal(t, len(defaultCategoryTable), len(actualCounts))

				for _, actualCount := range actualCounts {
					require.Contains(t, defaultCategoryTable, actualCount.CategoryId)
					expectedCount := expectedCountTable[actualCount.CategoryId]

					require.Equal(t, expectedCount, actualCount.Count)
				}
			}
		})

		t.Run("With hidden items", func(t *testing.T) {
			t.Run("Not including hidden items", func(t *testing.T) {
				countTables := []map[defs.Id]int64{
					{},
					{
						CategoryId_Clothing50_56: 1,
					},
					{
						CategoryId_Clothing50_56: 2,
					},
					{
						CategoryId_Clothing50_56: 2,
						CategoryId_Toys:          3,
					},
					{
						CategoryId_Clothing50_56:      1,
						CategoryId_Clothing56_62:      2,
						CategoryId_Clothing68_80:      3,
						CategoryId_Clothing86_92:      4,
						CategoryId_Clothing92_98:      5,
						CategoryId_Clothing104_116:    6,
						CategoryId_Clothing122_128:    7,
						CategoryId_Clothing128_140:    8,
						CategoryId_Clothing140_152:    9,
						CategoryId_Shoes:              10,
						CategoryId_Toys:               11,
						CategoryId_BabyChildEquipment: 12,
					},
				}

				for _, countTable := range countTables {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()

					for categoryId, count := range countTable {
						for i := int64(0); i < count; i++ {
							setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithItemCategory(categoryId), aux.WithHidden(true))
						}
					}

					actualCounts, err := queries.GetCategoryCounts(db, false)
					require.NoError(t, err)
					require.Equal(t, len(defaultCategoryTable), len(actualCounts))

					for _, count := range actualCounts {
						require.Contains(t, defaultCategoryTable, count.CategoryId)

						expectedCount := int64(0)
						actualCount := count.Count
						require.Equal(t, expectedCount, actualCount)
					}
				}
			})

			t.Run("Including hidden items", func(t *testing.T) {
				countTables := []map[defs.Id]int64{
					{},
					{
						defs.Clothing50_56: 1,
					},
					{
						defs.Clothing50_56: 2,
					},
					{
						defs.Clothing50_56: 2,
						defs.Toys:          3,
					},
					{
						defs.Clothing50_56:      1,
						defs.Clothing56_62:      2,
						defs.Clothing68_80:      3,
						defs.Clothing86_92:      4,
						defs.Clothing92_98:      5,
						defs.Clothing104_116:    6,
						defs.Clothing122_128:    7,
						defs.Clothing128_140:    8,
						defs.Clothing140_152:    9,
						defs.Shoes:              10,
						defs.Toys:               11,
						defs.BabyChildEquipment: 12,
					},
				}

				for _, countTable := range countTables {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()

					for categoryId, count := range countTable {
						for i := int64(0); i < count; i++ {
							setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithItemCategory(categoryId), aux.WithHidden(true))
						}
					}

					counts, err := queries.GetCategoryCounts(db, true)
					require.NoError(t, err)
					require.Equal(t, len(defs.ListCategoryIds()), len(counts))

					for _, count := range counts {
						require.Contains(t, defs.ListCategoryIds(), count.CategoryId)

						expectedCount := countTable[count.CategoryId]
						actualCount := count.Count
						require.Equal(t, expectedCount, actualCount)
					}
				}
			})
		})
	})
}
