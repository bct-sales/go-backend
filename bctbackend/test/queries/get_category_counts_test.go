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
	t.Run("Success", func(t *testing.T) {
		t.Run("Without hidden items", func(t *testing.T) {
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
						setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithItemCategory(categoryId), aux.WithHidden(false))
					}
				}

				counts, err := queries.GetCategoryCounts(db, true)
				require.NoError(t, err)
				require.Equal(t, len(defs.ListCategories()), len(counts))

				for _, count := range counts {
					require.Contains(t, defs.ListCategories(), count.CategoryId)

					expectedCount := countTable[count.CategoryId]
					actualCount := count.Count
					require.Equal(t, expectedCount, actualCount)
				}
			}
		})

		t.Run("With hidden items", func(t *testing.T) {
			t.Run("Not including hidden items", func(t *testing.T) {
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

					counts, err := queries.GetCategoryCounts(db, false)
					require.NoError(t, err)
					require.Equal(t, len(defs.ListCategories()), len(counts))

					for _, count := range counts {
						require.Contains(t, defs.ListCategories(), count.CategoryId)

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
					require.Equal(t, len(defs.ListCategories()), len(counts))

					for _, count := range counts {
						require.Contains(t, defs.ListCategories(), count.CategoryId)

						expectedCount := countTable[count.CategoryId]
						actualCount := count.Count
						require.Equal(t, expectedCount, actualCount)
					}
				}
			})
		})
	})
}
