//go:build test

package rest

import (
	"net/http"
	"testing"

	"bctbackend/algorithms"
	models "bctbackend/database/models"
	"bctbackend/rest"
	"bctbackend/rest/path"
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
	})
}
