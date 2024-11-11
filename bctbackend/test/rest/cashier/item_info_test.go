//go:build test

package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"bctbackend/database/models"
	restapi "bctbackend/rest/cashier"

	"bctbackend/test"

	"github.com/stretchr/testify/assert"
)

func TestGetItemInformation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, sale_count := range []int{0, 1, 2, 5} {
			label := fmt.Sprintf("Sale count: %d", sale_count)

			t.Run(label, func(t *testing.T) {
				sale_count := 0
				db, router := test.CreateRestRouter()
				writer := httptest.NewRecorder()
				defer db.Close()

				seller := test.AddSellerToDatabase(db)
				cashier := test.AddCashierToDatabase(db)
				item := test.AddItemToDatabase(db, seller.UserId, 1)

				for i := 0; i < sale_count; i++ {
					test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item.ItemId})
				}

				sessionId := test.AddSessionToDatabase(db, cashier.UserId)

				url := fmt.Sprintf("/api/v1/sales/items/%d", item.ItemId)
				request, err := http.NewRequest("GET", url, nil)

				if !assert.NoError(t, err) {
					request.AddCookie(test.CreateCookie(sessionId))
					router.ServeHTTP(writer, request)

					if assert.Equal(t, http.StatusCreated, writer.Code) {
						response := test.FromJson[restapi.GetItemInformationResponse](writer.Body.String())
						expectedHasBeenSold := sale_count > 0

						assert.Equal(t, item.Description, response.Description)
						assert.Equal(t, item.PriceInCents, response.PriceInCents)
						assert.Equal(t, item.CategoryId, response.CategoryId)
						assert.Equal(t, &expectedHasBeenSold, *response.HasBeenSold)
					}
				}
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Invalid URI", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			cashier := test.AddCashierToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, cashier.UserId)

			url := "/api/v1/sales/items/abc"
			request, err := http.NewRequest("GET", url, nil)

			if !assert.NoError(t, err) {
				request.AddCookie(test.CreateCookie(sessionId))
				router.ServeHTTP(writer, request)

				assert.Equal(t, http.StatusBadRequest, writer.Code)
			}
		})

		t.Run("As seller", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := test.AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)
			item := test.AddItemToDatabase(db, seller.UserId, 1)

			test.AddItemToDatabase(db, seller.UserId, 1)

			url := fmt.Sprintf("/api/v1/sales/items/%d", item.ItemId)
			request, err := http.NewRequest("GET", url, nil)

			if !assert.NoError(t, err) {
				request.AddCookie(test.CreateCookie(sessionId))
				router.ServeHTTP(writer, request)

				if assert.Equal(t, http.StatusForbidden, writer.Code) {
				}
			}
		})
	})
}
