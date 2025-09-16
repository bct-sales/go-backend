//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoesSellerHaveFrozenItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No frozen items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))

			result, err := queries.DoesSellerHaveFrozenItems(db, seller.UserId)
			require.NoError(t, err)
			require.False(t, result)
		})

		t.Run("Seller has one frozen item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(seller.UserId, 1, aux.WithFrozen(true), aux.WithHidden(false))

			result, err := queries.DoesSellerHaveFrozenItems(db, seller.UserId)
			require.NoError(t, err)
			require.True(t, result)
		})

		t.Run("Seller has nothing but frozen items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Items(seller.UserId, 10, aux.WithFrozen(true), aux.WithHidden(false))

			result, err := queries.DoesSellerHaveFrozenItems(db, seller.UserId)
			require.NoError(t, err)
			require.True(t, result)
		})

		t.Run("Other seller has frozen items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			otherSeller := setup.Seller()
			setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(otherSeller.UserId, 10, aux.WithFrozen(true), aux.WithHidden(false))

			result, err := queries.DoesSellerHaveFrozenItems(db, seller.UserId)
			require.NoError(t, err)
			require.False(t, result)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			invalidSellerId := models.Id(1)
			setup.RequireNoSuchUsers(t, invalidSellerId)

			_, err := queries.DoesSellerHaveFrozenItems(db, invalidSellerId)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})
	})
}
