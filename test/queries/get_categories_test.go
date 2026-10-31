//go:build test

package queries

import (
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetCategories(t *testing.T) {
	t.Run("1/2/3", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t)
		defer setup.Close()

		setup.Category(t, 1, "Alpha")
		setup.Category(t, 2, "Beta")
		setup.Category(t, 3, "Gamma")

		actualCategories, err := queries.GetCategories(t.Context(), db)
		require.NoError(t, err)
		require.Equal(t, 3, len(actualCategories))
		require.Equal(t, "Alpha", actualCategories[0].Name)
		require.Equal(t, "Beta", actualCategories[1].Name)
		require.Equal(t, "Gamma", actualCategories[2].Name)
	})

	t.Run("3/2/1", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t)
		defer setup.Close()

		setup.Category(t, 3, "Alpha")
		setup.Category(t, 2, "Beta")
		setup.Category(t, 1, "Gamma")

		actualCategories, err := queries.GetCategories(t.Context(), db)
		require.NoError(t, err)
		require.Equal(t, 3, len(actualCategories))
		require.Equal(t, "Gamma", actualCategories[0].Name)
		require.Equal(t, "Beta", actualCategories[1].Name)
		require.Equal(t, "Alpha", actualCategories[2].Name)
	})
}
