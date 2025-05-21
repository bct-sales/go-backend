//go:build test

package queries

import (
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"maps"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestCategoryWithIdExists(t *testing.T) {
	setup, db := NewDatabaseFixture(WithDefaultCategories)
	defer setup.Close()

	for categoryId := range maps.Keys(aux.DefaultCategoryTable()) {
		t.Run(fmt.Sprintf("categoryId = %d", categoryId), func(t *testing.T) {
			categoryExists, err := queries.CategoryWithIdExists(db, categoryId)

			require.NoError(t, err)
			require.True(t, categoryExists)
		})
	}
}
