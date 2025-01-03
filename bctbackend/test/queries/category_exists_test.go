//go:build test

package queries

import (
	"bctbackend/database/queries"
	"bctbackend/defs"
	"bctbackend/test"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestCategoryWithIdExists(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	for _, categoryId := range defs.ListCategories() {
		t.Run(fmt.Sprintf("categoryId = %d", categoryId), func(t *testing.T) {
			categoryExists, err := queries.CategoryWithIdExists(db, categoryId)

			if !assert.NoError(t, err) {
				return
			}

			if !assert.True(t, categoryExists) {
				return
			}
		})
	}
}
