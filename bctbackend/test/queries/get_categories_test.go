//go:build test

package queries

import (
	"bctbackend/database/queries"
	"bctbackend/defs"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetCategories(t *testing.T) {
	setup, db := NewDatabaseFixture(WithDefaultCategories)
	defer setup.Close()

	categories, err := queries.GetCategories(db)

	require.NoError(t, err)
	require.Equal(t, len(defs.ListCategoryIds()), len(categories))

	for _, category := range categories {
		require.Contains(t, defs.ListCategoryIds(), category.CategoryId)
	}
}
