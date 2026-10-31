//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoesSellerHaveFrozenItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No frozen items", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			setup.Items(t, seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				result, err := queries.DoesSellerHaveFrozenItems(t.Context(), db, seller.UserID)
				require.NoError(t, err)
				require.False(t, result)
			})
		})

		t.Run("Seller has one frozen item", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			setup.Items(t, seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(t, seller.UserID, 1, aux.WithFrozen(true), aux.WithHidden(false))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				result, err := queries.DoesSellerHaveFrozenItems(t.Context(), db, seller.UserID)
				require.NoError(t, err)
				require.True(t, result)
			})
		})

		t.Run("Seller has nothing but frozen items", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			setup.Items(t, seller.UserID, 10, aux.WithFrozen(true), aux.WithHidden(false))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				result, err := queries.DoesSellerHaveFrozenItems(t.Context(), db, seller.UserID)
				require.NoError(t, err)
				require.True(t, result)
			})
		})

		t.Run("Other sellers has frozen items", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			otherSeller1 := setup.Seller(t)
			seller := setup.Seller(t)
			otherSeller2 := setup.Seller(t)
			setup.Items(t, otherSeller1.UserID, 10, aux.WithFrozen(true), aux.WithHidden(false))
			setup.Items(t, seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(t, otherSeller2.UserID, 10, aux.WithFrozen(true), aux.WithHidden(false))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				result, err := queries.DoesSellerHaveFrozenItems(t.Context(), db, seller.UserID)
				require.NoError(t, err)
				require.False(t, result)
			})
		})

		t.Run("Seller has frozen items, other sellers don't", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			otherSeller1 := setup.Seller(t)
			seller := setup.Seller(t)
			otherSeller2 := setup.Seller(t)
			setup.Items(t, otherSeller1.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(t, seller.UserID, 10, aux.WithFrozen(true), aux.WithHidden(false))
			setup.Items(t, otherSeller2.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				result, err := queries.DoesSellerHaveFrozenItems(t.Context(), db, seller.UserID)
				require.NoError(t, err)
				require.True(t, result)
			})
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			setup.Seller(t)
			invalidSellerID := setup.GenerateNonexistentUserID(t)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				_, err := queries.DoesSellerHaveFrozenItems(t.Context(), db, invalidSellerID)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
			})
		})

		t.Run("Use on admin", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			user := setup.Admin(t)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				_, err := queries.DoesSellerHaveFrozenItems(t.Context(), db, user.UserID)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		})

		t.Run("Use on cashier", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			user := setup.Cashier(t)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				_, err := queries.DoesSellerHaveFrozenItems(t.Context(), db, user.UserID)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		})
	})
}
