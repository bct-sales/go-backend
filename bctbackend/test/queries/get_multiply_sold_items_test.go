//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetMultiplySoldItems(t *testing.T) {
	t.Run("No sales", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		test.AddItemToDatabase(db, sellerId, 1)

		multiplySoldItems, err := queries.GetMultiplySoldItems(db)

		require.NoError(t, err)
		require.Len(t, multiplySoldItems, 0)
	})

	t.Run("No multiply sold items", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		cashierId := test.AddCashierToDatabase(db).UserId

		itemIds := []models.Id{
			test.AddItemToDatabase(db, sellerId, 1).ItemId,
		}

		test.AddSaleToDatabase(db, cashierId, []models.Id{itemIds[0]})

		multiplySoldItems, err := queries.GetMultiplySoldItems(db)

		require.NoError(t, err)
		require.Len(t, multiplySoldItems, 0)
	})

	t.Run("Item sold twice", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		cashierId := test.AddCashierToDatabase(db).UserId

		items := []*models.Item{
			test.AddItemToDatabase(db, sellerId, 1),
		}

		sale1 := test.AddSaleToDatabase(db, cashierId, []models.Id{items[0].ItemId})
		sale2 := test.AddSaleToDatabase(db, cashierId, []models.Id{items[0].ItemId})

		multiplySoldItems, err := queries.GetMultiplySoldItems(db)

		require.NoError(t, err)
		require.Len(t, multiplySoldItems, 1)

		multiplySoldItem := multiplySoldItems[0]
		require.Equal(t, *(items[0]), multiplySoldItem.Item)
		require.Len(t, multiplySoldItem.Sales, 2)
		require.Equal(t, sale1, multiplySoldItem.Sales[0].SaleId)
		require.Equal(t, sale2, multiplySoldItem.Sales[1].SaleId)
	})
}
