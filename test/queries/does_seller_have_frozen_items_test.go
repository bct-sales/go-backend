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
		// t.Run("No frozen items", func(t *testing.T) {
		// 	setup, _ := NewDatabaseFixture(WithDefaultCategories)
		// 	defer setup.Close()

		// 	seller := setup.Seller()
		// 	setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))

		// 	setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
		// 		result, err := queries.DoesSellerHaveFrozenItems(db, seller.UserId)
		// 		require.NoError(t, err)
		// 		require.False(t, result)
		// 	})
		// })

		t.Run("Seller has one frozen item", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(seller.UserId, 1, aux.WithFrozen(true), aux.WithHidden(false))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				result, err := queries.DoesSellerHaveFrozenItems(db, seller.UserId)
				require.NoError(t, err)
				require.True(t, result)
			})
		})

		// t.Run("Seller has nothing but frozen items", func(t *testing.T) {
		// 	setup, _ := NewDatabaseFixture(WithDefaultCategories)
		// 	defer setup.Close()

		// 	seller := setup.Seller()
		// 	setup.Items(seller.UserId, 10, aux.WithFrozen(true), aux.WithHidden(false))

		// 	setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
		// 		result, err := queries.DoesSellerHaveFrozenItems(db, seller.UserId)
		// 		require.NoError(t, err)
		// 		require.True(t, result)
		// 	})
		// })

		// t.Run("Other sellers has frozen items", func(t *testing.T) {
		// 	setup, _ := NewDatabaseFixture(WithDefaultCategories)
		// 	defer setup.Close()

		// 	otherSeller1 := setup.Seller()
		// 	seller := setup.Seller()
		// 	otherSeller2 := setup.Seller()
		// 	setup.Items(otherSeller1.UserId, 10, aux.WithFrozen(true), aux.WithHidden(false))
		// 	setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
		// 	setup.Items(otherSeller2.UserId, 10, aux.WithFrozen(true), aux.WithHidden(false))

		// 	setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
		// 		result, err := queries.DoesSellerHaveFrozenItems(db, seller.UserId)
		// 		require.NoError(t, err)
		// 		require.False(t, result)
		// 	})
		// })

		// t.Run("Seller has frozen items, other sellers don't", func(t *testing.T) {
		// 	setup, _ := NewDatabaseFixture(WithDefaultCategories)
		// 	defer setup.Close()

		// 	otherSeller1 := setup.Seller()
		// 	seller := setup.Seller()
		// 	otherSeller2 := setup.Seller()
		// 	setup.Items(otherSeller1.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
		// 	setup.Items(seller.UserId, 10, aux.WithFrozen(true), aux.WithHidden(false))
		// 	setup.Items(otherSeller2.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))

		// 	setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
		// 		result, err := queries.DoesSellerHaveFrozenItems(db, seller.UserId)
		// 		require.NoError(t, err)
		// 		require.False(t, result)
		// 	})
		// })
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			setup.Seller()
			invalidSellerId := models.Id(999)
			setup.RequireNoSuchUsers(t, invalidSellerId)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				_, err := queries.DoesSellerHaveFrozenItems(db, invalidSellerId)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
			})
		})

		t.Run("Use on admin", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			user := setup.Admin()

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				_, err := queries.DoesSellerHaveFrozenItems(db, user.UserId)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		})

		t.Run("Use on cashier", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			user := setup.Cashier()

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				_, err := queries.DoesSellerHaveFrozenItems(db, user.UserId)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		})
	})
}
