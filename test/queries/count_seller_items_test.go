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

func TestCountSellerItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Single seller", func(t *testing.T) {
			for _, itemCount := range []int{0, 1, 2, 10, 100} {
				testLabel := fmt.Sprintf("Seller with %d items", itemCount)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()
					setup.Items(seller.UserID, itemCount, aux.WithHidden(false))

					actual, err := queries.CountSellerItems(db, seller.UserID, queries.IncludeAll, queries.IncludeAll)
					require.NoError(t, err)
					require.Equal(t, itemCount, actual)
				})
			}
		})

		t.Run("Multiple sellers", func(t *testing.T) {
			for _, itemCount := range []int{0, 1, 2, 10, 100} {
				testLabel := fmt.Sprintf("Seller with %d items", itemCount)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()
					otherSeller := setup.Seller()
					setup.Items(seller.UserID, itemCount, aux.WithHidden(false))
					setup.Items(otherSeller.UserID, itemCount, aux.WithHidden(false))

					actual, err := queries.CountSellerItems(db, seller.UserID, queries.IncludeAll, queries.IncludeAll)
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
				{queries.IncludeAll, queries.Exclude, baseCount + frozenCount},
				{queries.Exclude, queries.IncludeAll, baseCount + hiddenCount},
				{queries.IncludeAll, queries.IncludeAll, baseCount + frozenCount + hiddenCount},
				{queries.IncludeOnly, queries.IncludeOnly, 0},
				{queries.IncludeOnly, queries.IncludeAll, frozenCount},
				{queries.IncludeAll, queries.IncludeOnly, hiddenCount},
			} {
				t.Run(fmt.Sprintf("Frozen: %v, Hidden: %v", testCase.frozen, testCase.hidden), func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()
					setup.Items(seller.UserID, baseCount, aux.WithFrozen(false), aux.WithHidden(false))
					setup.Items(seller.UserID, frozenCount, aux.WithFrozen(true), aux.WithHidden(false))
					setup.Items(seller.UserID, hiddenCount, aux.WithFrozen(false), aux.WithHidden(true))

					actual, err := queries.CountSellerItems(db, seller.UserID, testCase.frozen, testCase.hidden)
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
			setup.Items(seller.UserID, baseCount, aux.WithFrozen(false), aux.WithHidden(false))
			setup.Items(seller.UserID, frozenCount, aux.WithFrozen(true), aux.WithHidden(false))
			setup.Items(seller.UserID, hiddenCount, aux.WithFrozen(false), aux.WithHidden(true))

			actual, err := queries.CountSellerItems(db, seller.UserID, queries.IncludeAll, queries.IncludeAll)
			require.NoError(t, err)
			require.Equal(t, baseCount+frozenCount+hiddenCount, actual)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No such user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			nonExistentSellerID := models.ID(1000)
			setup.RequireNoSuchUsers(t, nonExistentSellerID)

			_, err := queries.CountSellerItems(db, nonExistentSellerID, queries.IncludeAll, queries.IncludeAll)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})

		t.Run("Count items of cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()

			_, err := queries.CountSellerItems(db, cashier.UserID, queries.IncludeAll, queries.IncludeAll)
			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})

		t.Run("Count items of admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin()

			_, err := queries.CountSellerItems(db, admin.UserID, queries.IncludeAll, queries.IncludeAll)
			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})
	})
}
