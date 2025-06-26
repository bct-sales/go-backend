//go:build test

package rest

import (
	"fmt"
	"net/http"
	"slices"
	"testing"

	"bctbackend/algorithms"
	models "bctbackend/database/models"
	"bctbackend/server/path"
	"bctbackend/server/rest"
	shared "bctbackend/server/shared"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestGetAllSales(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Single sale", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))

			sale := setup.Sale(cashier.UserId, []models.Id{items[0].ItemID, items[1].ItemID})

			url := path.Sales().String()
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			actual := FromJson[rest.ListSalesSuccessResponse](t, writer.Body.String())
			expected := &rest.ListSalesSuccessResponse{
				Sales: []*rest.ListSalesSaleData{
					{
						SaleID:            sale.SaleID,
						CashierID:         cashier.UserId,
						TransactionTime:   shared.ConvertTimestampToDateTime(sale.TransactionTime),
						ItemCount:         2,
						TotalPriceInCents: items[0].PriceInCents + items[1].PriceInCents,
					},
				},
				SaleCount:      1,
				TotalSaleValue: items[0].PriceInCents + items[1].PriceInCents,
				ItemCount:      5,
				SoldItemCount:  2,
			}
			require.Equal(t, expected, actual)
		})

		t.Run("Two sales", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))

			sale1 := setup.Sale(cashier.UserId, []models.Id{items[0].ItemID, items[1].ItemID})
			sale2 := setup.Sale(cashier.UserId, []models.Id{items[2].ItemID, items[3].ItemID, items[4].ItemID})

			url := path.Sales().String()
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			actual := FromJson[rest.ListSalesSuccessResponse](t, writer.Body.String())
			expected := &rest.ListSalesSuccessResponse{
				Sales: []*rest.ListSalesSaleData{
					{
						SaleID:            sale1.SaleID,
						CashierID:         cashier.UserId,
						TransactionTime:   shared.ConvertTimestampToDateTime(sale1.TransactionTime),
						ItemCount:         2,
						TotalPriceInCents: items[0].PriceInCents + items[1].PriceInCents,
					},
					{
						SaleID:            sale2.SaleID,
						CashierID:         cashier.UserId,
						TransactionTime:   shared.ConvertTimestampToDateTime(sale2.TransactionTime),
						ItemCount:         3,
						TotalPriceInCents: items[2].PriceInCents + items[3].PriceInCents + items[4].PriceInCents,
					},
				},
				SaleCount:      2,
				TotalSaleValue: items[0].PriceInCents + items[1].PriceInCents + items[2].PriceInCents + items[3].PriceInCents + items[4].PriceInCents,
				ItemCount:      5,
				SoldItemCount:  5,
			}
			require.Equal(t, expected, actual)
		})

		t.Run("List all sales with startId", func(t *testing.T) {
			for _, k := range []int{1, 2, 5, 25} {
				testLabel := fmt.Sprintf("k = %d", k)
				t.Run(testLabel, func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					_, sessionId := setup.LoggedIn(setup.Admin())
					seller := setup.Seller()
					cashier := setup.Cashier()
					items := setup.Items(seller.UserId, 100, aux.WithHidden(false))

					for _, item := range items {
						setup.Sale(cashier.UserId, []models.Id{item.ItemID})
					}

					url := path.Sales().StartId(models.Id(k)).String()
					request := CreateGetRequest(url, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code)

					response := FromJson[rest.ListSalesSuccessResponse](t, writer.Body.String())
					expectedSaleCount := len(items) - k + 1
					require.Len(t, response.Sales, expectedSaleCount)
					require.Equal(t, 100, response.ItemCount)
					require.Equal(t, 100, response.SaleCount)
					require.Equal(t, 100, response.SoldItemCount)
				})
			}
		})

		t.Run("With limit and offset", func(t *testing.T) {
			for _, limit := range []int{1, 2, 5, 10} {
				for _, offset := range []int{0, 1, 2, 5, 10} {
					testLabel := fmt.Sprintf("limit = %d, offset = %d", limit, offset)
					t.Run(testLabel, func(t *testing.T) {
						setup, router, writer := NewRestFixture(WithDefaultCategories)
						defer setup.Close()

						_, sessionId := setup.LoggedIn(setup.Admin())
						seller := setup.Seller()
						cashier := setup.Cashier()
						items := setup.Items(seller.UserId, 100, aux.WithHidden(false))
						sales := algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserId, []models.Id{item.ItemID}) })

						url := path.Sales().Limit(limit).Offset(offset).String()
						request := CreateGetRequest(url, WithSessionCookie(sessionId))
						router.ServeHTTP(writer, request)
						require.Equal(t, http.StatusOK, writer.Code)

						response := FromJson[rest.ListSalesSuccessResponse](t, writer.Body.String())
						actualSales := response.Sales
						expectedSales := sales[offset : offset+limit]
						require.Len(t, actualSales, limit)
						for i, sale := range actualSales {
							require.Equal(t, expectedSales[i].SaleID, sale.SaleID)
						}
						require.Equal(t, 100, response.ItemCount)
						require.Equal(t, 100, response.SaleCount)
						require.Equal(t, 100, response.SoldItemCount)
					})
				}
			}
		})

		t.Run("With limit and offset, anti chronologically", func(t *testing.T) {
			for _, limit := range []int{1, 2, 5, 10} {
				for _, offset := range []int{0, 1, 2, 5, 10} {
					testLabel := fmt.Sprintf("limit = %d, offset = %d", limit, offset)
					t.Run(testLabel, func(t *testing.T) {
						setup, router, writer := NewRestFixture(WithDefaultCategories)
						defer setup.Close()

						_, sessionId := setup.LoggedIn(setup.Admin())
						seller := setup.Seller()
						cashier := setup.Cashier()
						items := setup.Items(seller.UserId, 100, aux.WithHidden(false))
						sales := algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserId, []models.Id{item.ItemID}) })

						url := path.Sales().Limit(limit).Offset(offset).AntiChronologically().String()
						request := CreateGetRequest(url, WithSessionCookie(sessionId))
						router.ServeHTTP(writer, request)
						require.Equal(t, http.StatusOK, writer.Code)

						response := FromJson[rest.ListSalesSuccessResponse](t, writer.Body.String())
						actualSales := response.Sales
						expectedSales := sales[:]
						slices.Reverse(expectedSales)
						expectedSales = expectedSales[offset : offset+limit]
						require.Len(t, actualSales, limit)
						for i, sale := range actualSales {
							require.Equal(t, expectedSales[i].SaleID, sale.SaleID)
						}
						require.Equal(t, 100, response.ItemCount)
						require.Equal(t, 100, response.SaleCount)
						require.Equal(t, 100, response.SoldItemCount)
					})
				}
			}
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("As seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{items[0].ItemID, items[1].ItemID})

			url := path.Sales().String()
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
		})

		t.Run("As cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier, sessionId := setup.LoggedIn(setup.Cashier())
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{items[0].ItemID, items[1].ItemID})

			url := path.Sales().String()
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
		})

		t.Run("No cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{items[0].ItemID, items[1].ItemID})

			url := path.Sales().String()
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Cookie with fake session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{items[0].ItemID, items[1].ItemID})

			url := path.Sales().String()
			request := CreateGetRequest(url, WithSessionCookie("fake_session_id"))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
		})

		t.Run("Cookie without session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{items[0].ItemID, items[1].ItemID})

			url := path.Sales().String()
			request := CreateGetRequest(url, WithCookie("whatever", "whatever"))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})
	})
}
