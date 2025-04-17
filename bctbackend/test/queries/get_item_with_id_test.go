//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetItemWithId(t *testing.T) {
	t.Run("Nonexisting item", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		itemId := models.NewId(1)
		_, err := queries.GetItemWithId(db, itemId)
		var NoSuchItemError *queries.NoSuchItemError
		require.ErrorAs(t, err, &NoSuchItemError)
	})

	t.Run("Existing item", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()
		item := setup.Item(seller.UserId, aux.WithDummyData(1))

		actual, err := queries.GetItemWithId(db, item.ItemId)
		require.NoError(t, err)
		require.Equal(t, item, actual)
	})
}
