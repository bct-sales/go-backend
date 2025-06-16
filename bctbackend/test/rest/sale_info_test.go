//go:build test

package rest

import (
	"net/http"
	"testing"

	"bctbackend/database/models"
	restapi "bctbackend/rest"
	"bctbackend/rest/path"
	rest "bctbackend/rest/shared"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestGetSaleInformation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("As admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()
			cashier := setup.Cashier()

			transactionTime := models.Timestamp(100)
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			saleId := setup.Sale(cashier.UserId, []models.Id{item.ItemID}, aux.WithTransactionTime(transactionTime))

			url := path.Sales().Id(saleId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			response := FromJson[restapi.GetSaleInformationSuccessResponse](t, writer.Body.String())
			require.Equal(t, cashier.UserId, response.CashierId)
			require.Equal(t, rest.ConvertTimestampToDateTime(transactionTime), response.TransactionTime)
			require.Equal(t, 1, len(response.Items))
			require.Equal(t, item.ItemID, response.Items[0].ItemId)
			require.Equal(t, item.SellerID, response.Items[0].SellerId)
			require.Equal(t, item.CategoryID, response.Items[0].CategoryId)
			require.Equal(t, item.Description, response.Items[0].Description)
			require.Equal(t, item.PriceInCents, response.Items[0].PriceInCents)
			require.Equal(t, item.Charity, *response.Items[0].Charity)
			require.Equal(t, item.Donation, *response.Items[0].Donation)
		})
	})
}
