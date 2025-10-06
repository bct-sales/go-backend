//go:build test

package queries

import (
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSoldItemsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No items in existence", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			query := queries.NewGetSoldItemsQuery()
			soldItems, err := query.Execute(db)
			require.NoError(t, err)
			require.Empty(t, soldItems)
		})
	})
}
