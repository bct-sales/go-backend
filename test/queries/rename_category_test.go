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

func TestRenameCategory(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Only one category in existence", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t)
			defer setup.Close()

			categoryID := models.ID(1)
			newName := "bar"
			setup.Category(t, categoryID, "foo")

			err := queries.RenameCategory(t.Context(), db, categoryID, newName)
			require.NoError(t, err)

			categories, err := queries.ListCategories(t.Context(), db)
			require.NoError(t, err)

			require.Len(t, categories, 1)
			require.Equal(t, categoryID, categories[0].CategoryID)
			require.Equal(t, newName, categories[0].Name)
		})

		t.Run("Multiple categories in existence", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t)
			defer setup.Close()

			setup.Category(t, 1, "foo")
			setup.Category(t, 2, "bar")
			setup.Category(t, 3, "baz")
			setup.Category(t, 4, "qux")

			err := queries.RenameCategory(t.Context(), db, 3, "bazzie")
			require.NoError(t, err)

			actualCategories, err := queries.ListCategories(t.Context(), db)
			require.NoError(t, err)

			require.Len(t, actualCategories, 4)

			expectedCategories := []*models.ItemCategory{
				{CategoryID: 1, Name: "foo"},
				{CategoryID: 2, Name: "bar"},
				{CategoryID: 3, Name: "bazzie"},
				{CategoryID: 4, Name: "qux"},
			}
			require.ElementsMatch(t, expectedCategories, actualCategories)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Invalid name", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t)
			defer setup.Close()

			setup.Category(t, 1, "foo")

			err := queries.RenameCategory(t.Context(), db, 1, "")
			requireDatabaseWrappedError(t, err, dberr.ErrInvalidCategoryName)
		})

		t.Run("Duplicate name", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t)
			defer setup.Close()

			setup.Category(t, 1, "foo")
			setup.Category(t, 2, "bar")

			err := queries.RenameCategory(t.Context(), db, 2, "foo")
			requireDatabaseWrappedError(t, err, dberr.ErrDuplicateCategoryName)
		})

		t.Run("Nonexistent category id", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t)
			defer setup.Close()

			err := queries.RenameCategory(t.Context(), db, 1, "foo")
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchCategory)
		})
	})
}
