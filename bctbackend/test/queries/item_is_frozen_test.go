//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestItemIsFrozen(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, frozen := range []bool{true, false} {
			t.Run("Frozen "+strconv.FormatBool(frozen), func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()
				item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithFrozen(frozen), aux.WithHidden(false))

				actual, err := queries.IsItemFrozen(db, item.ItemId)
				require.NoError(t, err)
				require.Equal(t, frozen, actual)
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No such item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			invalidId := models.Id(1)

			_, err := queries.IsItemFrozen(db, invalidId)
			var noSuchItemError *queries.NoSuchItemError
			require.ErrorAs(t, err, &noSuchItemError)
		})
	})
}
