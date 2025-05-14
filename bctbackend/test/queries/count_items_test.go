//go:build test

package queries

import (
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestCountItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, count := range []int{0, 1, 2, 5, 10, 23} {
			testLabel := fmt.Sprintf("%d items", count)
			t.Run(testLabel, func(t *testing.T) {
				setup, db := NewDatabaseFixture()
				defer setup.Close()

				seller := setup.Seller()
				setup.Items(seller.UserId, count)

				actual, err := queries.CountItems(db)
				require.NoError(t, err)
				require.Equal(t, count, actual)
			})
		}
	})
}
