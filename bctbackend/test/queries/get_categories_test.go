//go:build test

package queries

import (
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetCategories(t *testing.T) {
	setup, db := NewDatabaseFixture()
	defer setup.Close()

	setup.Category(1, "Alpha")
	setup.Category(2, "Beta")
	setup.Category(3, "Gamma")

	actualCategories, err := queries.GetCategories(db)
	require.NoError(t, err)
	require.Equal(t, 3, len(actualCategories))
	require.Equal(t, "Alpha", actualCategories[0].Name)
	require.Equal(t, "Beta", actualCategories[1].Name)
	require.Equal(t, "Gamma", actualCategories[2].Name)
}
