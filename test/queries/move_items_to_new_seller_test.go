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
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller(t)
			newSeller := setup.Seller(t)
			items := setup.Items(t, oldSeller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(t.Context(), db, oldSeller.UserID, newSeller.UserID)
				require.NoError(t, err)
			})

			for _, item := range items {
				newItem, err := queries.GetItemWithID(t.Context(), db, item.ItemID)
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
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller(t)
			newSeller := setup.Seller(t)
			items := setup.Items(t, oldSeller.UserID, 10, aux.WithFrozen(true), aux.WithHidden(false))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(t.Context(), db, oldSeller.UserID, newSeller.UserID)
				require.NoError(t, err)
			})

			for _, item := range items {
				newItem, err := queries.GetItemWithID(t.Context(), db, item.ItemID)
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
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller(t)
			newSeller := setup.Seller(t)
			items := setup.Items(t, oldSeller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(true))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(t.Context(), db, oldSeller.UserID, newSeller.UserID)
				require.NoError(t, err)
			})

			for _, item := range items {
				newItem, err := queries.GetItemWithID(t.Context(), db, item.ItemID)
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
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			otherSeller1 := setup.Seller(t)
			oldSeller := setup.Seller(t)
			otherSeller2 := setup.Seller(t)
			newSeller := setup.Seller(t)
			otherSeller3 := setup.Seller(t)

			setup.Items(t, otherSeller1.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(t, oldSeller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(t, otherSeller2.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(t, newSeller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(t, otherSeller3.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))

			var itemsBefore []*models.Item
			query := queries.NewGetItemsQuery()
			require.NoError(t, query.Execute(t.Context(), db, queries.CollectTo(&itemsBefore)))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(t.Context(), db, oldSeller.UserID, newSeller.UserID)
				require.NoError(t, err)
			})

			for _, itemBefore := range itemsBefore {
				itemAfter, err := queries.GetItemWithID(t.Context(), db, itemBefore.ItemID)
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
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			newSeller := setup.Seller(t)
			invalidSellerID := setup.GenerateNonexistentUserID(t)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(t.Context(), db, invalidSellerID, newSeller.UserID)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
			})
		})

		t.Run("Nonexistent new seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller(t)
			invalidSellerID := setup.GenerateNonexistentUserID(t)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(t.Context(), db, oldSeller.UserID, invalidSellerID)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
			})
		})

		t.Run("Admin old seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Admin(t)
			newSeller := setup.Seller(t)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(t.Context(), db, oldSeller.UserID, newSeller.UserID)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		})

		t.Run("Cashier old seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Cashier(t)
			newSeller := setup.Seller(t)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(t.Context(), db, oldSeller.UserID, newSeller.UserID)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		})

		t.Run("Admin new seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller(t)
			newSeller := setup.Admin(t)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(t.Context(), db, oldSeller.UserID, newSeller.UserID)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		})

		t.Run("Cashier new seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller(t)
			newSeller := setup.Cashier(t)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(t.Context(), db, oldSeller.UserID, newSeller.UserID)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			})
		})
	})
}
