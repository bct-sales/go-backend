//go:build test

package rest

import (
	"net/http"
	"testing"

	"bctbackend/algorithms"
	models "bctbackend/database/models"
	path "bctbackend/server/paths"
	"bctbackend/server/rest"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestListCashierSales(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Cashier views own sales", func(t *testing.T) {
			t.Run("All sales", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()
				cashier, sessionID := setup.LoggedIn(setup.Cashier())
				cashier2 := setup.Cashier()

				items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
				sales := algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserID, []models.ID{item.ItemID}) })
				items2 := setup.Items(seller.UserID, 20, aux.WithHidden(false))
				algorithms.Map(items2, func(item *models.Item) *models.Sale { return setup.Sale(cashier2.UserID, []models.ID{item.ItemID}) })

				url := path.CashierSales(cashier.UserID)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				actual := FromJSON[rest.ListCashierSalesSuccessResponse](t, writer.Body.String())
				require.NotNil(t, actual)
				require.Equal(t, len(sales), len(actual.Sales))
				require.Equal(t, len(sales), actual.SaleCount)
			})

			t.Run("With offset", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()
				cashier, sessionID := setup.LoggedIn(setup.Cashier())

				items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
				sales := algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserID, []models.ID{item.ItemID}) })

				url := path.CashierSales(cashier.UserID).AddOffset(1)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				actual := FromJSON[rest.ListCashierSalesSuccessResponse](t, writer.Body.String())
				require.NotNil(t, actual)
				require.Equal(t, len(sales)-1, len(actual.Sales))
				require.Equal(t, sales[1].SaleID, actual.Sales[0].SaleID)
				require.Equal(t, len(sales), actual.SaleCount)
			})

			t.Run("With limit", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()
				cashier, sessionID := setup.LoggedIn(setup.Cashier())

				items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
				sales := algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserID, []models.ID{item.ItemID}) })

				url := path.CashierSales(cashier.UserID).AddLimit(1)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				actual := FromJSON[rest.ListCashierSalesSuccessResponse](t, writer.Body.String())
				require.NotNil(t, actual)
				require.Equal(t, 1, len(actual.Sales))
				require.Equal(t, sales[0].SaleID, actual.Sales[0].SaleID)
				require.Equal(t, len(sales), actual.SaleCount)
			})

			t.Run("With limit and offset", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()
				cashier, sessionID := setup.LoggedIn(setup.Cashier())

				items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
				sales := algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserID, []models.ID{item.ItemID}) })

				url := path.CashierSales(cashier.UserID).AddLimit(3).AddOffset(2)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				actual := FromJSON[rest.ListCashierSalesSuccessResponse](t, writer.Body.String())
				require.NotNil(t, actual)
				require.Equal(t, 3, len(actual.Sales))
				require.Equal(t, sales[2].SaleID, actual.Sales[0].SaleID)
				require.Equal(t, len(sales), actual.SaleCount)
			})

			t.Run("Anti chronologically", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()
				cashier, sessionID := setup.LoggedIn(setup.Cashier())

				items := setup.Items(seller.UserID, 3, aux.WithHidden(false))
				sales := algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserID, []models.ID{item.ItemID}) })

				url := path.CashierSales(cashier.UserID).AddAntiChronologicalOrder()
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				actual := FromJSON[rest.ListCashierSalesSuccessResponse](t, writer.Body.String())
				require.NotNil(t, actual)
				require.Equal(t, 3, len(actual.Sales))
				require.Equal(t, sales[2].SaleID, actual.Sales[0].SaleID)
				require.Equal(t, sales[1].SaleID, actual.Sales[1].SaleID)
				require.Equal(t, sales[0].SaleID, actual.Sales[2].SaleID)
				require.Equal(t, len(sales), actual.SaleCount)
			})
		})

		t.Run("Admin views sales", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionID := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()
			cashier := setup.Cashier()
			cashier2 := setup.Cashier()

			items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
			sales := algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserID, []models.ID{item.ItemID}) })
			items2 := setup.Items(seller.UserID, 20, aux.WithHidden(false))
			algorithms.Map(items2, func(item *models.Item) *models.Sale { return setup.Sale(cashier2.UserID, []models.ID{item.ItemID}) })

			url := path.CashierSales(cashier.UserID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			actual := FromJSON[rest.ListCashierSalesSuccessResponse](t, writer.Body.String())
			require.NotNil(t, actual)
			require.Equal(t, len(sales), len(actual.Sales))
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Seller views sales", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionID := setup.LoggedIn(setup.Seller())
			cashier := setup.Cashier()
			cashier2 := setup.Cashier()

			items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
			algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserID, []models.ID{item.ItemID}) })
			items2 := setup.Items(seller.UserID, 20, aux.WithHidden(false))
			algorithms.Map(items2, func(item *models.Item) *models.Sale { return setup.Sale(cashier2.UserID, []models.ID{item.ItemID}) })

			url := path.CashierSales(cashier.UserID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusForbidden, writer.Code)
		})

		t.Run("Other cashier views sales", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			cashier2, sessionID := setup.LoggedIn(setup.Cashier())

			items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
			algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserID, []models.ID{item.ItemID}) })
			items2 := setup.Items(seller.UserID, 20, aux.WithHidden(false))
			algorithms.Map(items2, func(item *models.Item) *models.Sale { return setup.Sale(cashier2.UserID, []models.ID{item.ItemID}) })

			url := path.CashierSales(cashier.UserID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusForbidden, writer.Code)
		})
	})
}
