//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveUserWithID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		{
			sellerExists, err := queries.UserWithIDExists(db, seller.UserID)
			require.NoError(t, err)
			require.True(t, sellerExists)
		}
		{
			cashierExists, err := queries.UserWithIDExists(db, cashier.UserID)
			require.NoError(t, err)
			require.True(t, cashierExists)
		}

		err := queries.RemoveUserWithID(db, seller.UserID)
		require.NoError(t, err)

		{
			sellerExists, err := queries.UserWithIDExists(db, seller.UserID)
			require.NoError(t, err)
			require.False(t, sellerExists)
		}
		{
			cashierExists, err := queries.UserWithIDExists(db, cashier.UserID)
			require.NoError(t, err)
			require.True(t, cashierExists)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Removing non-existing user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			userID := models.ID(99999)
			setup.RequireNoSuchUsers(t, userID)

			err := queries.RemoveUserWithID(db, userID)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})
	})
}
