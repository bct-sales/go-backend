//go:build test

package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/database/queries"
	restapi "bctbackend/rest/cashier"

	"bctbackend/test"

	"github.com/stretchr/testify/assert"
)

func TestAddSaleItem(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	seller := test.AddSellerToDatabase(db)
	cashier := test.AddCashierToDatabase(db)
	item := test.AddItemToDatabase(db, seller.UserId, 1)
	sessionId := test.AddSessionToDatabase(db, cashier.UserId)
	payload := restapi.AddSalePayload{
		CashierId: cashier.UserId,
		Items:     []models.Id{item.ItemId},
	}
	url := "/api/v1/sales"
	request := test.CreatePostRequest(url, &payload)
	request.AddCookie(test.CreateCookie(sessionId))

	router.ServeHTTP(writer, request)

	if assert.Equal(t, http.StatusCreated, writer.Code) {
		response := test.FromJson[restapi.AddSaleResponse](writer.Body.String())

		sale, err := queries.GetSaleWithId(db, response.SaleId)
		if assert.NoError(t, err) {
			assert.Equal(t, cashier.UserId, sale.CashierId)

			saleItems, err := queries.GetSaleItems(db, sale.SaleId)
			if assert.NoError(t, err) {
				assert.Len(t, saleItems, 1)
				assert.Equal(t, item.ItemId, saleItems[0].ItemId)
			}
		}
	}
}
