//go:build test

package queries

import (
	"bctbackend/database/queries"
	"bctbackend/defs"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestCategoryWithIdExists(t *testing.T) {
	setup, db := NewDatabaseFixture()
	defer setup.Close()

	for _, categoryId := range defs.ListCategories() {
		t.Run(fmt.Sprintf("categoryId = %d", categoryId), func(t *testing.T) {
			categoryExists, err := queries.CategoryWithIdExists(db, categoryId)

			require.NoError(t, err)
			require.True(t, categoryExists)
		})
	}
}
