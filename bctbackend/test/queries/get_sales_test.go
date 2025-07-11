//go:build test

package queries

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
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

		t.Run("Get sales with limit and offset", func(t *testing.T) {
			for _, limit := range []int{1, 2, 5, 10} {
				for _, offset := range []int{0, 1, 2, 5, 10} {
					testLabel := fmt.Sprintf("limit = %d, offset = %d", limit, offset)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()
						cashier := setup.Cashier()

						items := setup.Items(seller.UserId, 100, aux.WithHidden(false))
						sales := algorithms.Map(items, func(item *models.Item) *models.Sale {
							return setup.Sale(cashier.UserId, []models.Id{item.ItemID})
						})

						expectedSales := sales[offset : offset+limit]
						actualSales := []*models.SaleSummary{}
						err := queries.NewGetSalesQuery().WithRowSelection(limit, offset).Execute(db, queries.CollectTo(&actualSales))
						require.NoError(t, err)
						require.Len(t, actualSales, limit)

						for index, actualSale := range actualSales {
							require.Equal(t, cashier.UserId, actualSale.CashierID)

							saleItems, err := queries.GetSaleItems(db, actualSale.SaleID)
							require.NoError(t, err)
							require.Equal(t, 1, len(saleItems))
							require.Equal(t, expectedSales[index].SaleID, actualSale.SaleID)
						}
					})
				}
			}
		})

		t.Run("Get sales with limit and offset, anti chronologically", func(t *testing.T) {
			for _, limit := range []int{1, 2, 5, 10} {
				for _, offset := range []int{0, 1, 2, 5, 10} {
					testLabel := fmt.Sprintf("limit = %d, offset = %d", limit, offset)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()
						cashier := setup.Cashier()

						items := setup.Items(seller.UserId, 100, aux.WithHidden(false))

						sales := []*models.Sale{}
						for index, item := range items {
							sale := setup.Sale(cashier.UserId, []models.Id{item.ItemID}, aux.WithTransactionTime(models.Timestamp(index)))
							sales = append(sales, sale)
						}

						expectedSales := sales[:]
						slices.Reverse(expectedSales)
						expectedSales = expectedSales[offset : offset+limit]
						actualSales := []*models.SaleSummary{}
						err := queries.NewGetSalesQuery().WithRowSelection(limit, offset).OrderedAntiChronologically().Execute(db, queries.CollectTo(&actualSales))
						require.NoError(t, err)
						require.Len(t, actualSales, limit)

						for index, actualSale := range actualSales {
							require.Equal(t, cashier.UserId, actualSale.CashierID)

							saleItems, err := queries.GetSaleItems(db, actualSale.SaleID)
							require.NoError(t, err)
							require.Equal(t, 1, len(saleItems))
							require.Equal(t, expectedSales[index].SaleID, actualSale.SaleID)
						}
					})
				}
			}
		})
	})
}
