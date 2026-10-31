//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureUserExists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		user := setup.Seller(t)

		err := queries.EnsureUserExists(t.Context(), db, user.UserID)
		require.NoError(t, err)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		nonexistentUserID := setup.GenerateNonexistentUserID(t)

		err := queries.EnsureUserExists(t.Context(), db, nonexistentUserID)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
	})
}
