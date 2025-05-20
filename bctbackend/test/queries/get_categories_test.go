//go:build test

package queries

import (
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetCategories(t *testing.T) {
	setup, db := NewDatabaseFixture(WithDefaultCategories)
	defer setup.Close()

	categoryTable := DefaultCategoryTable()

	actualCategories, err := queries.GetCategories(db)
	require.NoError(t, err)
	require.Equal(t, len(categoryTable), len(actualCategories))

	for _, actualCategory := range actualCategories {
		expectedCategoryName, ok := categoryTable[actualCategory.CategoryId]
		require.True(t, ok)
		require.Equal(t, expectedCategoryName, actualCategory.Name)
	}
}
