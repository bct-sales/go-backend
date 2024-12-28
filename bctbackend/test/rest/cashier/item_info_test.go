//go:build test

package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"bctbackend/database/models"
	"bctbackend/database/queries"
	restapi "bctbackend/rest/cashier"
	"bctbackend/rest/path"

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

				url := path.SalesItems().WithItemId(item.ItemId)
				request := test.CreateGetRequest(url)

				request.AddCookie(test.CreateCookie(sessionId))
				router.ServeHTTP(writer, request)

				if assert.Equal(t, http.StatusOK, writer.Code) {
					response := test.FromJson[restapi.GetItemInformationSuccessResponse](writer.Body.String())
					expectedHasBeenSold := sale_count > 0

					assert.Equal(t, item.Description, response.Description)
					assert.Equal(t, item.PriceInCents, response.PriceInCents)
					assert.Equal(t, item.CategoryId, response.CategoryId)
					assert.Equal(t, expectedHasBeenSold, *response.HasBeenSold)
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

			url := path.SalesItems().WithRawItemId("abc")
			request := test.CreateGetRequest(url)
			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusBadRequest, writer.Code)
		})

		t.Run("As seller", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := test.AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)
			item := test.AddItemToDatabase(db, seller.UserId, 1)

			test.AddItemToDatabase(db, seller.UserId, 1)

			url := path.SalesItems().WithItemId(item.ItemId)
			request := test.CreateGetRequest(url)
			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusForbidden, writer.Code)
		})

		t.Run("As admin", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			admin := test.AddAdminToDatabase(db)
			seller := test.AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, admin.UserId)
			item := test.AddItemToDatabase(db, seller.UserId, 1)

			test.AddItemToDatabase(db, seller.UserId, 1)

			url := path.SalesItems().WithItemId(item.ItemId)
			request := test.CreateGetRequest(url)
			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusForbidden, writer.Code)
		})

		t.Run("Nonexistent item", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			admin := test.AddAdminToDatabase(db)
			seller := test.AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, admin.UserId)
			nonexistentItem := models.NewId(1)

			itemExists, err := queries.ItemWithIdExists(db, nonexistentItem)

			if !assert.NoError(t, err) {
				if assert.False(t, itemExists) {
					test.AddItemToDatabase(db, seller.UserId, 1)

					url := path.SalesItems().WithItemId(nonexistentItem)
					request := test.CreateGetRequest(url)

					request.AddCookie(test.CreateCookie(sessionId))
					router.ServeHTTP(writer, request)

					assert.Equal(t, http.StatusBadRequest, writer.Code)
				}
			}
		})
	})
}
