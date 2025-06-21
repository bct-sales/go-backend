//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSales(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Get all sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()

			items := setup.Items(seller.UserId, 100, aux.WithHidden(false))

			for _, item := range items {
				setup.Sale(cashier.UserId, []models.Id{item.ItemID})
			}

			actualSales := []*models.SaleSummary{}
			err := queries.NewGetSalesQuery().Execute(db, queries.CollectTo(&actualSales))
			require.NoError(t, err)
			require.Len(t, actualSales, len(items))

			for _, actualSale := range actualSales {
				require.Equal(t, cashier.UserId, actualSale.CashierID)

				saleItems, err := queries.GetSaleItems(db, actualSale.SaleID)
				require.NoError(t, err)
				require.Equal(t, 1, len(saleItems))
			}
		})

		t.Run("Get sales with id higher than", func(t *testing.T) {
			for k := range 10 {
				testLabel := fmt.Sprintf("k = %d", k)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()
					cashier := setup.Cashier()

					items := setup.Items(seller.UserId, 100, aux.WithHidden(false))

					saleIds := make([]models.Id, len(items))
					for _, item := range items {
						setup.Sale(cashier.UserId, []models.Id{item.ItemID})
					}

					actualSales := []*models.SaleSummary{}
					err := queries.NewGetSalesQuery().WithIdGreaterThanOrEqualTo(models.Id(k+1)).Execute(db, queries.CollectTo(&actualSales))
					require.NoError(t, err)
					require.Len(t, actualSales, len(saleIds)-k)

					for _, actualSale := range actualSales {
						require.Equal(t, cashier.UserId, actualSale.CashierID)

						saleItems, err := queries.GetSaleItems(db, actualSale.SaleID)
						require.NoError(t, err)
						require.Equal(t, 1, len(saleItems))
					}
				})
			}
		})
	})
}
