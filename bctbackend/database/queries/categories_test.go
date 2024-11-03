package queries

import (
	models "bctbackend/database/models"
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
		assert.Equal(t, len(models.Categories()), len(categories))

		for _, category := range categories {
			assert.Contains(t, models.Categories(), category.CategoryId)
		}
	}
}

func TestCategoryWithIdExists(t *testing.T) {
	db := openInitializedDatabase()
	defer db.Close()

	for _, categoryId := range models.Categories() {
		t.Run(fmt.Sprintf("categoryId = %d", categoryId), func(t *testing.T) {
			assert.True(t, CategoryWithIdExists(db, categoryId))
		})
	}
}

func TestGetCategoryCounts(t *testing.T) {
	countTables := []map[models.Id]int64{
		{},
		{
			models.Clothing50_56: 1,
		},
		{
			models.Clothing50_56: 2,
		},
		{
			models.Clothing50_56: 2,
			models.Toys:          3,
		},
		{
			models.Clothing50_56:      1,
			models.Clothing56_62:      2,
			models.Clothing68_80:      3,
			models.Clothing86_92:      4,
			models.Clothing92_98:      5,
			models.Clothing104_116:    6,
			models.Clothing122_128:    7,
			models.Clothing128_140:    8,
			models.Clothing140_152:    9,
			models.Shoes:              10,
			models.Toys:               11,
			models.BabyChildEquipment: 12,
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
			assert.Equal(t, len(models.Categories()), len(counts))

			for _, count := range counts {
				assert.Contains(t, models.Categories(), count.CategoryId)

				expectedCount := countTable[count.CategoryId]
				actualCount := count.Count
				assert.Equal(t, expectedCount, actualCount)
			}
		}
	}
}
