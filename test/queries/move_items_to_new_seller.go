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
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			oldSeller := setup.Seller()
			newSeller := setup.Seller()
			setup.Items(oldSeller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, oldSeller.UserId, newSeller.UserId)
				require.NoError(t, err)
			})
		})

	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent old seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			invalidSellerId := models.Id(1)
			setup.RequireNoSuchUsers(t, invalidSellerId)
			newSeller := setup.Seller()

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.MoveItemsToNewSeller(db, invalidSellerId, newSeller.UserId)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
			})
		})
	})
}
