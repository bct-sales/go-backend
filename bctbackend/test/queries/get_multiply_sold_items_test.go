//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetMultiplySoldItems(t *testing.T) {
	t.Run("No sales", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		AddItemToDatabase(db, sellerId, WithDummyData(1))

		multiplySoldItems, err := queries.GetMultiplySoldItems(db)

		require.NoError(t, err)
		require.Len(t, multiplySoldItems, 0)
	})

	t.Run("No multiply sold items", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		cashierId := AddCashierToDatabase(db).UserId

		itemIds := []models.Id{
			AddItemToDatabase(db, sellerId, WithDummyData(1)).ItemId,
		}

		AddSaleToDatabase(db, cashierId, []models.Id{itemIds[0]})

		multiplySoldItems, err := queries.GetMultiplySoldItems(db)

		require.NoError(t, err)
		require.Len(t, multiplySoldItems, 0)
	})

	t.Run("Item sold twice", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		cashierId := AddCashierToDatabase(db).UserId

		items := []*models.Item{
			AddItemToDatabase(db, sellerId, WithDummyData(1)),
		}

		sale1 := AddSaleToDatabase(db, cashierId, []models.Id{items[0].ItemId})
		sale2 := AddSaleToDatabase(db, cashierId, []models.Id{items[0].ItemId})

		multiplySoldItems, err := queries.GetMultiplySoldItems(db)

		require.NoError(t, err)
		require.Len(t, multiplySoldItems, 1)

		multiplySoldItem := multiplySoldItems[0]
		require.Equal(t, *(items[0]), multiplySoldItem.Item)
		require.Len(t, multiplySoldItem.Sales, 2)
		require.Equal(t, sale1, multiplySoldItem.Sales[0].SaleId)
		require.Equal(t, sale2, multiplySoldItem.Sales[1].SaleId)
	})

	t.Run("Item sold thrice", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		cashierId := AddCashierToDatabase(db).UserId

		items := []*models.Item{
			AddItemToDatabase(db, sellerId, WithDummyData(1)),
		}

		sale1 := AddSaleToDatabase(db, cashierId, []models.Id{items[0].ItemId})
		sale2 := AddSaleToDatabase(db, cashierId, []models.Id{items[0].ItemId})
		sale3 := AddSaleToDatabase(db, cashierId, []models.Id{items[0].ItemId})

		multiplySoldItems, err := queries.GetMultiplySoldItems(db)

		require.NoError(t, err)
		require.Len(t, multiplySoldItems, 1)

		multiplySoldItem := multiplySoldItems[0]
		require.Equal(t, *(items[0]), multiplySoldItem.Item)
		require.Len(t, multiplySoldItem.Sales, 3)
		require.Equal(t, sale1, multiplySoldItem.Sales[0].SaleId)
		require.Equal(t, sale2, multiplySoldItem.Sales[1].SaleId)
		require.Equal(t, sale3, multiplySoldItem.Sales[2].SaleId)
	})

	t.Run("Two items sold twice", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		cashierId := AddCashierToDatabase(db).UserId

		items := []*models.Item{
			AddItemToDatabase(db, sellerId, WithDummyData(1)),
			AddItemToDatabase(db, sellerId, WithDummyData(2)),
		}

		sale1 := AddSaleToDatabase(db, cashierId, []models.Id{items[0].ItemId, items[1].ItemId})
		sale2 := AddSaleToDatabase(db, cashierId, []models.Id{items[0].ItemId, items[1].ItemId})

		multiplySoldItems, err := queries.GetMultiplySoldItems(db)

		require.NoError(t, err)
		require.Len(t, multiplySoldItems, 2)

		require.Equal(t, *(items[0]), multiplySoldItems[0].Item)
		require.Len(t, multiplySoldItems[0].Sales, 2)
		require.Equal(t, sale1, multiplySoldItems[0].Sales[0].SaleId)
		require.Equal(t, sale2, multiplySoldItems[0].Sales[1].SaleId)

		require.Equal(t, *(items[1]), multiplySoldItems[1].Item)
		require.Len(t, multiplySoldItems[1].Sales, 2)
		require.Equal(t, sale1, multiplySoldItems[1].Sales[0].SaleId)
		require.Equal(t, sale2, multiplySoldItems[1].Sales[1].SaleId)
	})

	t.Run("Sales [1], [1, 2], [1, 2, 3]", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		cashierId := AddCashierToDatabase(db).UserId

		items := []*models.Item{
			AddItemToDatabase(db, sellerId, WithDummyData(1)),
			AddItemToDatabase(db, sellerId, WithDummyData(2)),
			AddItemToDatabase(db, sellerId, WithDummyData(3)),
		}

		sale1 := AddSaleToDatabase(db, cashierId, []models.Id{items[0].ItemId})
		sale2 := AddSaleToDatabase(db, cashierId, []models.Id{items[0].ItemId, items[1].ItemId})
		sale3 := AddSaleToDatabase(db, cashierId, []models.Id{items[0].ItemId, items[1].ItemId, items[2].ItemId})

		multiplySoldItems, err := queries.GetMultiplySoldItems(db)

		require.NoError(t, err)
		require.Len(t, multiplySoldItems, 2)

		require.Equal(t, *(items[0]), multiplySoldItems[0].Item)
		require.Len(t, multiplySoldItems[0].Sales, 3)
		require.Equal(t, sale1, multiplySoldItems[0].Sales[0].SaleId)
		require.Equal(t, sale2, multiplySoldItems[0].Sales[1].SaleId)
		require.Equal(t, sale3, multiplySoldItems[0].Sales[2].SaleId)

		require.Equal(t, *(items[1]), multiplySoldItems[1].Item)
		require.Len(t, multiplySoldItems[1].Sales, 2)
		require.Equal(t, sale2, multiplySoldItems[1].Sales[0].SaleId)
		require.Equal(t, sale3, multiplySoldItems[1].Sales[1].SaleId)
	})
}
