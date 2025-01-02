//go:build test

package queries

import (
	"bctbackend/database/queries"
	"bctbackend/defs"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestGetCategories(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	categories, err := queries.GetCategories(db)
	if assert.NoError(t, err) {
		assert.Equal(t, len(defs.ListCategories()), len(categories))

		for _, category := range categories {
			assert.Contains(t, defs.ListCategories(), category.CategoryId)
		}
	}
}
