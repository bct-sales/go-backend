//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCategoryWithIdExists(t *testing.T) {
	setup, db := NewDatabaseFixture()
	defer setup.Close()

	setup.Category(1, "Good")
	setup.Category(2, "Bad")
	setup.Category(3, "Ugly")

	for i := models.Id(1); i <= 3; i++ {
		categoryExists, err := queries.CategoryWithIdExists(db, i)

		require.NoError(t, err)
		require.True(t, categoryExists)
	}

	for i := models.Id(4); i <= 10; i++ {
		categoryExists, err := queries.CategoryWithIdExists(db, i)

		require.NoError(t, err)
		require.False(t, categoryExists)
	}
}
