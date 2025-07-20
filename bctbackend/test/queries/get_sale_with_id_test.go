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

func TestGetSaleWithId(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()
		item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
		sale := setup.Sale(cashier.UserId, []models.Id{item.ItemID})

		actual, err := queries.GetSaleWithId(db, sale.SaleID)
		require.NoError(t, err)
		require.NotNil(t, actual)
		require.Equal(t, sale, actual)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		saleId := models.Id(999)
		setup.RequireNoSuchSales(t, saleId)

		_, err := queries.GetSaleWithId(db, saleId)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSale)
	})
}
