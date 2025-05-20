//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"

	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAddCategory(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Note the lack of WithDefaultCategories here
		// We want to start with a clean slate
		setup, db := NewDatabaseFixture()
		defer setup.Close()

		categoryName := "Test Category"
		id := models.Id(1)
		err := queries.AddCategory(db, models.Id(1), categoryName)
		require.NoError(t, err, `Failed to add category: %v`, err)

		categoryExists, err := queries.CategoryWithIdExists(db, id)
		require.True(t, categoryExists)
		require.NoError(t, err)

		table, err := queries.GetCategoryMap(db)
		require.NoError(t, err)
		actual, ok := table[id]
		require.True(t, ok)
		require.Equal(t, categoryName, actual)
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Invalid name", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			id := models.Id(1)
			categoryName := ""
			err := queries.AddCategory(db, id, categoryName)
			var invalidCategoryNameError *queries.InvalidCategoryNameError
			require.ErrorAs(t, err, &invalidCategoryNameError)
		})
	})
}
