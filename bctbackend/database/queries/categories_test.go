package queries

import (
	"bctbackend/defs"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestGetCategories(t *testing.T) {
	db := openInitializedDatabase()
	defer db.Close()

	categories, err := GetCategories(db)
	if assert.NoError(t, err) {
		assert.Equal(t, len(defs.ListCategories()), len(categories))

		for _, category := range categories {
			assert.Contains(t, defs.ListCategories(), category.CategoryId)
		}
	}
}

func TestCategoryWithIdExists(t *testing.T) {
	db := openInitializedDatabase()
	defer db.Close()

	for _, categoryId := range defs.ListCategories() {
		t.Run(fmt.Sprintf("categoryId = %d", categoryId), func(t *testing.T) {
			assert.True(t, CategoryWithIdExists(db, categoryId))
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
		db := openInitializedDatabase()
		defer db.Close()

		sellerId := addTestSeller(db)

		for categoryId, count := range countTable {
			for i := int64(0); i < count; i++ {
				addTestItemInCategory(db, sellerId, categoryId)
			}
		}

		counts, err := GetCategoryCounts(db)
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
