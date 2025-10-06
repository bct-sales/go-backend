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

func TestGetSaleWithID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()
		item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
		sale := setup.Sale(cashier.UserID, []models.ID{item.ItemID})

		actual, err := queries.GetSaleWithID(db, sale.SaleID)
		require.NoError(t, err)
		require.NotNil(t, actual)
		require.Equal(t, sale, actual)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		nonexistentSaleID := setup.GenerateNonexistentSaleID(t)

		_, err := queries.GetSaleWithID(db, nonexistentSaleID)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSale)
	})
}
