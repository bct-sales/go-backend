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
	t.Run("zero items", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		counts, err := GetCategoryCounts(db)
		if assert.NoError(t, err) {
			assert.Equal(t, len(models.Categories()), len(counts))

			for _, count := range counts {
				assert.Contains(t, models.Categories(), count.CategoryId)
				assert.Equal(t, int64(0), count.Count)
			}
		}
	})

	t.Run("multiple items", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		countTable := map[models.Id]int64{
			models.Clothing50_56:      2,
			models.Toys:               3,
			models.BabyChildEquipment: 5,
		}

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
	})
}
