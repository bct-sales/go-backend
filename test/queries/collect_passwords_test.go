//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollectPasswords(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Zero passwords", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			actual, err := queries.CollectPasswords(t.Context(), db)
			expected := []string{}
			require.NoError(t, err)
			require.ElementsMatch(t, expected, actual)
		})

		t.Run("One password", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			_, err := queries.AddUser(t.Context(), db, models.NewAdminRoleID(), 0, nil, "a")
			require.NoError(t, err)

			actual, err := queries.CollectPasswords(t.Context(), db)
			expected := []string{"a"}
			require.NoError(t, err)
			require.ElementsMatch(t, expected, actual)
		})

		t.Run("Multiple passwords", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			_, err := queries.AddUser(t.Context(), db, models.NewAdminRoleID(), 0, nil, "a")
			require.NoError(t, err)
			_, err = queries.AddUser(t.Context(), db, models.NewCashierRoleID(), 0, nil, "b")
			require.NoError(t, err)
			_, err = queries.AddUser(t.Context(), db, models.NewSellerRoleID(), 0, nil, "c")
			require.NoError(t, err)

			actual, err := queries.CollectPasswords(t.Context(), db)
			expected := []string{"a", "b", "c"}
			require.NoError(t, err)
			require.ElementsMatch(t, expected, actual)
		})
	})
}
