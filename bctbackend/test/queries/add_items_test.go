//go:build test

package queries

import (
	"bctbackend/database/queries"
	. "bctbackend/test/setup"

	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Add zero items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			callback := func(addItem queries.AddItemFunction) {}

			err := queries.AddItems(db, callback)
			require.NoError(t, err)

			count, err := queries.CountItems(db, queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 0, count)
		})
	})
}
