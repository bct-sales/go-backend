//go:build test

package queries

import (
	"bctbackend/database/queries"
	"bctbackend/defs"
	"bctbackend/test"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestGetCategories(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	categories, err := queries.GetCategories(db)
	if assert.NoError(t, err) {
		assert.Equal(t, len(defs.ListCategories()), len(categories))

		for _, category := range categories {
			assert.Contains(t, defs.ListCategories(), category.CategoryId)
		}
	}
}

func TestCategoryWithIdExists(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	for _, categoryId := range defs.ListCategories() {
		t.Run(fmt.Sprintf("categoryId = %d", categoryId), func(t *testing.T) {
			assert.True(t, queries.CategoryWithIdExists(db, categoryId))
		})
	}
}

func TestGetCategoryCounts(t *testing.T) {
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
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId

		for categoryId, count := range countTable {
			for i := int64(0); i < count; i++ {
				test.AddItemInCategoryToDatabase(db, sellerId, categoryId)
			}
		}

		counts, err := queries.GetCategoryCounts(db)
		if assert.NoError(t, err) {
			assert.Equal(t, len(defs.ListCategories()), len(counts))

			for _, count := range counts {
				assert.Contains(t, defs.ListCategories(), count.CategoryId)

				expectedCount := countTable[count.CategoryId]
				actualCount := count.Count
				assert.Equal(t, expectedCount, actualCount)
			}
		}
	}
}
