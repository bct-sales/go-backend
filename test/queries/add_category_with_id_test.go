//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"

	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddCategoryWithID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Note the lack of WithDefaultCategories here
		// We want to start with a clean slate
		setup, db := NewDatabaseFixture(t)
		defer setup.Close()

		categoryName := "Test Category"
		id := models.ID(1)
		err := queries.AddCategoryWithID(t.Context(), db, models.ID(1), categoryName)
		require.NoError(t, err, `Failed to add category: %v`, err)

		categoryExists, err := queries.CategoryWithIDExists(t.Context(), db, id)
		require.True(t, categoryExists)
		require.NoError(t, err)

		table, err := queries.GetCategoryNameTable(t.Context(), db)
		require.NoError(t, err)
		actual, ok := table[id]
		require.True(t, ok)
		require.Equal(t, categoryName, actual)
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Invalid name", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t)
			defer setup.Close()

			id := models.ID(1)
			categoryName := ""
			err := queries.AddCategoryWithID(t.Context(), db, id, categoryName)
			requireDatabaseWrappedError(t, err, dberr.ErrInvalidCategoryName)
		})

		t.Run("Duplicate name", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t)
			defer setup.Close()

			categoryName := "Foo"

			{
				id := models.ID(1)
				err := queries.AddCategoryWithID(t.Context(), db, id, categoryName)
				require.NoError(t, err)
			}

			{
				id := models.ID(2)
				err := queries.AddCategoryWithID(t.Context(), db, id, categoryName)
				requireDatabaseWrappedError(t, err, dberr.ErrDuplicateCategoryName)
			}
		})

		t.Run("Same id used twice", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t)
			defer setup.Close()

			setup.Category(t, 1, "Test Category")

			id := models.ID(1)
			categoryName := "xyz"
			err := queries.AddCategoryWithID(t.Context(), db, id, categoryName)
			requireDatabaseWrappedError(t, err, dberr.ErrIDAlreadyInUse)
		})
	})
}
