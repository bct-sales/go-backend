//go:build test

package queries

import (
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestCountItems(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		count, err := queries.CountItems(db)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, 0, count) {
			return
		}
	})

	t.Run("One item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		test.AddItemToDatabase(db, sellerId, 1)

		count, err := queries.CountItems(db)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, 1, count) {
			return
		}
	})

	t.Run("Two items", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		test.AddItemToDatabase(db, sellerId, 1)
		test.AddItemToDatabase(db, sellerId, 2)

		count, err := queries.CountItems(db)
		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, 2, count) {
			return
		}
	})
}
