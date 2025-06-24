//go:build test

package rest

import (
	"net/http"
	"testing"

	"bctbackend/algorithms"
	models "bctbackend/database/models"
	"bctbackend/server/path"
	"bctbackend/server/rest"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestListCashierSales(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Cashier views own sales", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier, sessionId := setup.LoggedIn(setup.Cashier())
			cashier2 := setup.Cashier()

			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			sales := algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserId, []models.Id{item.ItemID}) })
			items2 := setup.Items(seller.UserId, 20, aux.WithHidden(false))
			algorithms.Map(items2, func(item *models.Item) *models.Sale { return setup.Sale(cashier2.UserId, []models.Id{item.ItemID}) })

			url := path.CashierSales().WithCashierId(cashier.UserId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			actual := FromJson[rest.GetCashierSalesSuccessResponse](t, writer.Body.String())
			require.NotNil(t, actual)
			require.Equal(t, len(sales), len(actual.Sales))
		})

		t.Run("Admin views sales", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()
			cashier := setup.Cashier()
			cashier2 := setup.Cashier()

			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			sales := algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserId, []models.Id{item.ItemID}) })
			items2 := setup.Items(seller.UserId, 20, aux.WithHidden(false))
			algorithms.Map(items2, func(item *models.Item) *models.Sale { return setup.Sale(cashier2.UserId, []models.Id{item.ItemID}) })

			url := path.CashierSales().WithCashierId(cashier.UserId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			actual := FromJson[rest.GetCashierSalesSuccessResponse](t, writer.Body.String())
			require.NotNil(t, actual)
			require.Equal(t, len(sales), len(actual.Sales))
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Seller views sales", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())
			cashier := setup.Cashier()
			cashier2 := setup.Cashier()

			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserId, []models.Id{item.ItemID}) })
			items2 := setup.Items(seller.UserId, 20, aux.WithHidden(false))
			algorithms.Map(items2, func(item *models.Item) *models.Sale { return setup.Sale(cashier2.UserId, []models.Id{item.ItemID}) })

			url := path.CashierSales().WithCashierId(cashier.UserId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusForbidden, writer.Code)
		})

		t.Run("Other cashier views sales", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			cashier2, sessionId := setup.LoggedIn(setup.Cashier())

			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			algorithms.Map(items, func(item *models.Item) *models.Sale { return setup.Sale(cashier.UserId, []models.Id{item.ItemID}) })
			items2 := setup.Items(seller.UserId, 20, aux.WithHidden(false))
			algorithms.Map(items2, func(item *models.Item) *models.Sale { return setup.Sale(cashier2.UserId, []models.Id{item.ItemID}) })

			url := path.CashierSales().WithCashierId(cashier.UserId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusForbidden, writer.Code)
		})
	})
}
