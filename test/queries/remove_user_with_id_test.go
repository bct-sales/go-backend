//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveUserWithID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		cashier := setup.Cashier(t)

		{
			sellerExists, err := queries.UserWithIDExists(t.Context(), db, seller.UserID)
			require.NoError(t, err)
			require.True(t, sellerExists)
		}
		{
			cashierExists, err := queries.UserWithIDExists(t.Context(), db, cashier.UserID)
			require.NoError(t, err)
			require.True(t, cashierExists)
		}

		err := queries.RemoveUserWithID(t.Context(), db, seller.UserID)
		require.NoError(t, err)

		{
			sellerExists, err := queries.UserWithIDExists(t.Context(), db, seller.UserID)
			require.NoError(t, err)
			require.False(t, sellerExists)
		}
		{
			cashierExists, err := queries.UserWithIDExists(t.Context(), db, cashier.UserID)
			require.NoError(t, err)
			require.True(t, cashierExists)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Removing non-existing user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			userID := setup.GenerateNonexistentUserID(t)

			err := queries.RemoveUserWithID(t.Context(), db, userID)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})
	})
}
