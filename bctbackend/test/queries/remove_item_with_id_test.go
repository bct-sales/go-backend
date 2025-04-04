//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test"
	aux "bctbackend/test/helpers"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestRemoveExistingItem(t *testing.T) {
	setup, db := Setup()
	defer setup.Close()

	seller := setup.Seller()
	itemId := setup.Item(seller.UserId, aux.WithDummyData(1)).ItemId

	err := queries.RemoveItemWithId(db, itemId)

	require.NoError(t, err)

	itemExists, err := queries.ItemWithIdExists(db, itemId)
	require.NoError(t, err)
	require.False(t, itemExists)
}

func TestRemoveNonexistingItem(t *testing.T) {
	setup, db := Setup()
	defer setup.Close()

	itemId := models.NewId(1)

	err := queries.RemoveItemWithId(db, itemId)

	var NoSuchItemError *queries.NoSuchItemError
	require.ErrorAs(t, err, &NoSuchItemError)
}

func TestRemoveSoldItem(t *testing.T) {
	setup, db := Setup()
	defer setup.Close()

	seller := setup.Seller()
	cashier := setup.Cashier()
	itemId := setup.Item(seller.UserId, aux.WithDummyData(1)).ItemId

	setup.Sale(cashier.UserId, []models.Id{itemId})

	err := queries.RemoveItemWithId(db, itemId)
	require.Error(t, err)

	itemExists, err := queries.ItemWithIdExists(db, itemId)
	require.NoError(t, err)
	require.True(t, itemExists)
}
