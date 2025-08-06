//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTotalSalesValue(t *testing.T) {
	t.Run("Zero sales", func(t *testing.T) {
		t.Run("Only visible items in count", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Items(seller.UserId, 0, aux.WithHidden(false))

			total, err := queries.GetTotalSalesValue(db)
			require.NoError(t, err)
			require.Equal(t, models.MoneyInCents(0), total)
		})
	})
}
