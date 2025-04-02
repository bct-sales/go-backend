//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestItemIsFrozen(t *testing.T) {
	for _, frozen := range []bool{true, false} {
		t.Run("Frozen "+strconv.FormatBool(frozen), func(t *testing.T) {
			db := OpenInitializedDatabase()
			defer db.Close()

			item := AddItemToDatabase(db, AddSellerToDatabase(db).UserId, WithDummyData(1), WithFrozen(frozen))

			actual, err := queries.ItemWithIdIsFrozen(db, item.ItemId)
			require.NoError(t, err)
			require.Equal(t, frozen, actual)
		})
	}

	t.Run("No such item", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		invalidId := models.Id(1)

		_, err := queries.ItemWithIdIsFrozen(db, invalidId)
		var noSuchItemError *queries.NoSuchItemError
		require.ErrorAs(t, err, &noSuchItemError)
	})
}
