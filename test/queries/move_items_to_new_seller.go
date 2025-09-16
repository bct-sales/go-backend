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

func TestMoveItemsToNewSeller(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No frozen items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller()
			newSeller := setup.Seller()
			items := setup.Items(oldSeller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, oldSeller.UserId, newSeller.UserId)
				require.NoError(t, err)
			})

			for _, item := range items {
				newItem, err := queries.GetItemWithId(db, item.ItemID)
				require.NoError(t, err)

				require.Equal(t, item.CategoryID, newItem.CategoryID)
				require.Equal(t, item.Charity, newItem.Charity)
				require.Equal(t, item.Description, newItem.Description)
				require.Equal(t, item.Donation, newItem.Donation)
				require.Equal(t, newSeller, newItem.SellerID)
				require.Equal(t, item.Frozen, newItem.Frozen)
				require.Equal(t, item.Hidden, newItem.Hidden)
				require.Equal(t, item.PriceInCents, newItem.PriceInCents)
			}
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent old seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			newSeller := setup.Seller()
			invalidSellerId := models.Id(999)
			setup.RequireNoSuchUsers(t, invalidSellerId)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, invalidSellerId, newSeller.UserId)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
			})
		})

		t.Run("Nonexistent new seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller()
			invalidSellerId := models.Id(999)
			setup.RequireNoSuchUsers(t, invalidSellerId)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, oldSeller.UserId, invalidSellerId)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
			})
		})
	})
}
