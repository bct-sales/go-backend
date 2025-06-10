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

func TestGetSellerItemCount(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Single seller", func(t *testing.T) {
			for _, itemCount := range []int{0, 1, 2, 10, 100} {
				testLabel := fmt.Sprintf("Seller with %d items", itemCount)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()
					setup.Items(seller.UserId, itemCount, aux.WithHidden(false))

					actual, err := queries.GetSellerItemCount(db, seller.UserId, queries.Include, queries.Include)
					require.NoError(t, err)
					require.Equal(t, itemCount, actual)
				})
			}
		})

		t.Run("Multiple sellers", func(t *testing.T) {
			for _, itemCount := range []int{0, 1, 2, 10, 100} {
				testLabel := fmt.Sprintf("Seller with %d items", itemCount)
				t.Run(testLabel, func(t *testing.T) {
					t.Parallel()

					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()
					otherSeller := setup.Seller()
					setup.Items(seller.UserId, itemCount, aux.WithHidden(false))
					setup.Items(otherSeller.UserId, itemCount, aux.WithHidden(false))

					actual, err := queries.GetSellerItemCount(db, seller.UserId, queries.Include, queries.Include)
					require.NoError(t, err)
					require.Equal(t, itemCount, actual)
				})
			}
		})

		t.Run("Flags", func(t *testing.T) {
			baseCount := 4
			frozenCount := 8
			hiddenCount := 16

			for _, testCase := range []struct {
				frozen   queries.GetSellerItemCountFlag
				hidden   queries.GetSellerItemCountFlag
				expected int
			}{
				{queries.Exclude, queries.Exclude, baseCount},
				{queries.Include, queries.Exclude, baseCount + frozenCount},
				{queries.Exclude, queries.Include, baseCount + hiddenCount},
				{queries.Include, queries.Include, baseCount + frozenCount + hiddenCount},
				{queries.Exclusive, queries.Exclusive, 0},
				{queries.Exclusive, queries.Include, frozenCount},
				{queries.Include, queries.Exclusive, hiddenCount},
			} {
				t.Run(fmt.Sprintf("Frozen: %v, Hidden: %v", testCase.frozen, testCase.hidden), func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()
					setup.Items(seller.UserId, baseCount, aux.WithFrozen(false), aux.WithHidden(false))
					setup.Items(seller.UserId, frozenCount, aux.WithFrozen(true), aux.WithHidden(false))
					setup.Items(seller.UserId, hiddenCount, aux.WithFrozen(false), aux.WithHidden(true))

					actual, err := queries.GetSellerItemCount(db, seller.UserId, testCase.frozen, testCase.hidden)
					require.NoError(t, err)
					require.Equal(t, testCase.expected, actual)
				})
			}
		})

		t.Run("Counting hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			baseCount := 4
			frozenCount := 8
			hiddenCount := 16
			setup.Items(seller.UserId, baseCount, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(seller.UserId, frozenCount, aux.WithFrozen(true), aux.WithHidden(false))
			setup.Items(seller.UserId, hiddenCount, aux.WithFrozen(false), aux.WithHidden(true))

			actual, err := queries.GetSellerItemCount(db, seller.UserId, queries.Include, queries.Include)
			require.NoError(t, err)
			require.Equal(t, baseCount+frozenCount+hiddenCount, actual)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No such user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			nonExistentSellerId := models.Id(1000)
			setup.RequireNoSuchUsers(t, nonExistentSellerId)

			_, err := queries.GetSellerItemCount(db, nonExistentSellerId, queries.Include, queries.Include)
			require.ErrorIs(t, err, dberr.ErrNoSuchUser)
		})

		t.Run("Count items of cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()

			_, err := queries.GetSellerItemCount(db, cashier.UserId, queries.Include, queries.Include)
			require.ErrorIs(t, err, dberr.ErrWrongRole)
		})

		t.Run("Count items of admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin()

			_, err := queries.GetSellerItemCount(db, admin.UserId, queries.Include, queries.Include)
			require.ErrorIs(t, err, dberr.ErrWrongRole)
		})
	})
}
