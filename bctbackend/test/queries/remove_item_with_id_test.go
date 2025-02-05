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

func TestRemoveExistingItem(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	sellerId := AddSellerToDatabase(db).UserId
	itemId := AddItemToDatabase(db, sellerId, WithDummyData(1)).ItemId

	err := queries.RemoveItemWithId(db, itemId)

	require.NoError(t, err)

	itemExists, err := queries.ItemWithIdExists(db, itemId)
	require.NoError(t, err)
	require.False(t, itemExists)
}

func TestRemoveNonexistingItem(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	itemId := models.NewId(1)

	err := queries.RemoveItemWithId(db, itemId)

	var NoSuchItemError *queries.NoSuchItemError
	require.ErrorAs(t, err, &NoSuchItemError)
}

func TestRemoveSoldItem(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	sellerId := AddSellerToDatabase(db).UserId
	cashierId := AddCashierToDatabase(db).UserId
	itemId := AddItemToDatabase(db, sellerId, WithDummyData(1)).ItemId

	AddSaleToDatabase(db, cashierId, []models.Id{itemId})

	err := queries.RemoveItemWithId(db, itemId)
	require.Error(t, err)

	itemExists, err := queries.ItemWithIdExists(db, itemId)
	require.NoError(t, err)
	require.True(t, itemExists)
}
