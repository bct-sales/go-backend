//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestItemIsHidden(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, hidden := range []bool{true, false} {
			testLabel := fmt.Sprintf("Hidden %t", hidden)
			t.Run(testLabel, func(t *testing.T) {
				t.Parallel()

				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()
				item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithFrozen(false), aux.WithHidden(hidden))

				actual, err := queries.IsItemHidden(db, item.ItemID)
				require.NoError(t, err)
				require.Equal(t, hidden, actual)
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			invalidId := models.Id(1)

			_, err := queries.IsItemHidden(db, invalidId)
			require.ErrorIs(t, err, dberr.ErrNoSuchItem)
		})
	})
}
