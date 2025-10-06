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
			items := setup.Items(oldSeller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, oldSeller.UserID, newSeller.UserID)
				require.NoError(t, err)
			})

			for _, item := range items {
				newItem, err := queries.GetItemWithID(db, item.ItemID)
				require.NoError(t, err)

				require.Equal(t, item.CategoryID, newItem.CategoryID)
				require.Equal(t, item.Charity, newItem.Charity)
				require.Equal(t, item.Description, newItem.Description)
				require.Equal(t, item.Donation, newItem.Donation)
				require.Equal(t, newSeller.UserID, newItem.SellerID)
				require.Equal(t, item.Frozen, newItem.Frozen)
				require.Equal(t, item.Hidden, newItem.Hidden)
				require.Equal(t, item.PriceInCents, newItem.PriceInCents)
			}
		})

		t.Run("Frozen items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller()
			newSeller := setup.Seller()
			items := setup.Items(oldSeller.UserID, 10, aux.WithFrozen(true), aux.WithHidden(false))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, oldSeller.UserID, newSeller.UserID)
				require.NoError(t, err)
			})

			for _, item := range items {
				newItem, err := queries.GetItemWithID(db, item.ItemID)
				require.NoError(t, err)

				require.Equal(t, item.CategoryID, newItem.CategoryID)
				require.Equal(t, item.Charity, newItem.Charity)
				require.Equal(t, item.Description, newItem.Description)
				require.Equal(t, item.Donation, newItem.Donation)
				require.Equal(t, newSeller.UserID, newItem.SellerID)
				require.Equal(t, item.Frozen, newItem.Frozen)
				require.Equal(t, item.Hidden, newItem.Hidden)
				require.Equal(t, item.PriceInCents, newItem.PriceInCents)
			}
		})

		t.Run("Hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller()
			newSeller := setup.Seller()
			items := setup.Items(oldSeller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(true))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, oldSeller.UserID, newSeller.UserID)
				require.NoError(t, err)
			})

			for _, item := range items {
				newItem, err := queries.GetItemWithID(db, item.ItemID)
				require.NoError(t, err)

				require.Equal(t, item.CategoryID, newItem.CategoryID)
				require.Equal(t, item.Charity, newItem.Charity)
				require.Equal(t, item.Description, newItem.Description)
				require.Equal(t, item.Donation, newItem.Donation)
				require.Equal(t, newSeller.UserID, newItem.SellerID)
				require.Equal(t, item.Frozen, newItem.Frozen)
				require.Equal(t, item.Hidden, newItem.Hidden)
				require.Equal(t, item.PriceInCents, newItem.PriceInCents)
			}
		})

		t.Run("Only seller's items will get moved", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			otherSeller1 := setup.Seller()
			oldSeller := setup.Seller()
			otherSeller2 := setup.Seller()
			newSeller := setup.Seller()
			otherSeller3 := setup.Seller()

			setup.Items(otherSeller1.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(oldSeller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(otherSeller2.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(newSeller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(otherSeller3.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))

			var itemsBefore []*models.Item
			query := queries.NewGetItemsQuery()
			require.NoError(t, query.Execute(db, queries.CollectTo(&itemsBefore)))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, oldSeller.UserID, newSeller.UserID)
				require.NoError(t, err)
			})

			for _, itemBefore := range itemsBefore {
				itemAfter, err := queries.GetItemWithID(db, itemBefore.ItemID)
				require.NoError(t, err)

				require.Equal(t, itemBefore.CategoryID, itemAfter.CategoryID)
				require.Equal(t, itemBefore.Charity, itemAfter.Charity)
				require.Equal(t, itemBefore.Description, itemAfter.Description)
				require.Equal(t, itemBefore.Donation, itemAfter.Donation)
				require.Equal(t, itemBefore.Frozen, itemAfter.Frozen)
				require.Equal(t, itemBefore.Hidden, itemAfter.Hidden)
				require.Equal(t, itemBefore.PriceInCents, itemAfter.PriceInCents)

				var expectedSeller models.ID
				if itemBefore.SellerID == oldSeller.UserID {
					expectedSeller = newSeller.UserID
				} else {
					expectedSeller = itemBefore.SellerID
				}

				require.Equal(t, expectedSeller, itemAfter.SellerID)
			}
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent old seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			newSeller := setup.Seller()
			invalidSellerID := setup.GenerateNonexistentUserID(t)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, invalidSellerID, newSeller.UserID)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
			})
		})

		t.Run("Nonexistent new seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller()
			invalidSellerID := setup.GenerateNonexistentUserID(t)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, oldSeller.UserID, invalidSellerID)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
			})
		})

		t.Run("Admin old seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Admin()
			newSeller := setup.Seller()

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, oldSeller.UserID, newSeller.UserID)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		})

		t.Run("Cashier old seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Cashier()
			newSeller := setup.Seller()

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, oldSeller.UserID, newSeller.UserID)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		})

		t.Run("Admin new seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller()
			newSeller := setup.Admin()

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, oldSeller.UserID, newSeller.UserID)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		})

		t.Run("Cashier new seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller()
			newSeller := setup.Cashier()

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, oldSeller.UserID, newSeller.UserID)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		})
	})
}
