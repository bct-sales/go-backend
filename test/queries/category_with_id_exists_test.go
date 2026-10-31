//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCategoryWithIDExists(t *testing.T) {
	setup, db := NewDatabaseFixture(t)
	defer setup.Close()

	setup.Category(t, 1, "Good")
	setup.Category(t, 2, "Bad")
	setup.Category(t, 3, "Ugly")

	for i := models.ID(1); i <= 3; i++ {
		categoryExists, err := queries.CategoryWithIDExists(t.Context(), db, i)

		require.NoError(t, err)
		require.True(t, categoryExists)
	}

	for i := models.ID(4); i <= 10; i++ {
		categoryExists, err := queries.CategoryWithIDExists(t.Context(), db, i)

		require.NoError(t, err)
		require.False(t, categoryExists)
	}
}
