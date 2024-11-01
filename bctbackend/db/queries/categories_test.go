package queries

import (
	models "bctbackend/db/models"
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
