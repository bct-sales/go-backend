//go:build test

package queries

import (
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestCountItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Not including hidden items in count", func(t *testing.T) {
			t.Run("No hidden items", func(t *testing.T) {
				for _, count := range []int{0, 1, 2, 5, 10, 23} {
					testLabel := fmt.Sprintf("%d unhidden items", count)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture()
						defer setup.Close()

						seller := setup.Seller()
						setup.Items(seller.UserId, count, aux.WithHidden(false))

						actual, err := queries.CountItems(db, false)
						require.NoError(t, err)
						require.Equal(t, count, actual)
					})
				}
			})

			t.Run("With hidden items", func(t *testing.T) {
				setup, db := NewDatabaseFixture()
				defer setup.Close()

				seller := setup.Seller()
				setup.Item(seller.UserId, aux.WithHidden(true))

				actual, err := queries.CountItems(db, false)
				require.NoError(t, err)
				require.Equal(t, 0, actual)
			})
		})

		t.Run("Including hidden items in count", func(t *testing.T) {
			t.Run("No hidden items", func(t *testing.T) {
				for _, count := range []int{0, 1, 2, 5, 10, 23} {
					testLabel := fmt.Sprintf("%d unhidden items", count)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture()
						defer setup.Close()

						seller := setup.Seller()
						setup.Items(seller.UserId, count, aux.WithHidden(false))

						actual, err := queries.CountItems(db, true)
						require.NoError(t, err)
						require.Equal(t, count, actual)
					})
				}
			})

			t.Run("With hidden items", func(t *testing.T) {
				setup, db := NewDatabaseFixture()
				defer setup.Close()

				seller := setup.Seller()
				setup.Item(seller.UserId, aux.WithHidden(true))

				actual, err := queries.CountItems(db, true)
				require.NoError(t, err)
				require.Equal(t, 1, actual)
			})
		})
	})
}
