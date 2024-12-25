//go:build test

package rest

import (
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

func TestAddSaleItem(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, router := test.CreateRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		seller := test.AddSellerToDatabase(db)
		cashier := test.AddCashierToDatabase(db)
		item := test.AddItemToDatabase(db, seller.UserId, 1)
		sessionId := test.AddSessionToDatabase(db, cashier.UserId)
		payload := restapi.AddSalePayload{
			Items: []models.Id{item.ItemId},
		}
		url := path.Sales().String()
		request := test.CreatePostRequest(url, &payload)
		request.AddCookie(test.CreateCookie(sessionId))

		router.ServeHTTP(writer, request)

		if assert.Equal(t, http.StatusCreated, writer.Code) {
			response := test.FromJson[restapi.AddSaleSuccessResponse](writer.Body.String())

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
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("As seller", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := test.AddSellerToDatabase(db)
			item := test.AddItemToDatabase(db, seller.UserId, 1)
			sessionId := test.AddSessionToDatabase(db, seller.UserId) // Causes the operation to fail
			payload := restapi.AddSalePayload{
				Items: []models.Id{item.ItemId},
			}
			url := path.Sales().String()
			request := test.CreatePostRequest(url, &payload)
			request.AddCookie(test.CreateCookie(sessionId))

			router.ServeHTTP(writer, request)

			if assert.Equal(t, http.StatusForbidden, writer.Code) {
				sales, err := queries.GetSales(db)

				if assert.NoError(t, err) {
					assert.Empty(t, sales)
				}
			}
		})

		t.Run("As admin", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			admin := test.AddAdminToDatabase(db)
			seller := test.AddSellerToDatabase(db)
			item := test.AddItemToDatabase(db, seller.UserId, 1)
			sessionId := test.AddSessionToDatabase(db, admin.UserId) // Causes the operation to fail
			payload := restapi.AddSalePayload{
				Items: []models.Id{item.ItemId},
			}
			url := path.Sales().String()
			request := test.CreatePostRequest(url, &payload)
			request.AddCookie(test.CreateCookie(sessionId))

			router.ServeHTTP(writer, request)

			if assert.Equal(t, http.StatusForbidden, writer.Code) {
				sales, err := queries.GetSales(db)

				if assert.NoError(t, err) {
					assert.Empty(t, sales)
				}
			}
		})

		t.Run("No items", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			cashier := test.AddCashierToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, cashier.UserId)
			payload := restapi.AddSalePayload{
				Items: []models.Id{},
			}
			url := path.Sales().String()
			request := test.CreatePostRequest(url, &payload)
			request.AddCookie(test.CreateCookie(sessionId))

			router.ServeHTTP(writer, request)

			if assert.Equal(t, http.StatusBadRequest, writer.Code) {
				sales, err := queries.GetSales(db)

				if assert.NoError(t, err) {
					assert.Empty(t, sales)
				}
			}
		})

		t.Run("Nonexistent item", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			cashier := test.AddCashierToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, cashier.UserId)
			nonexistentItemId := models.Id(1000)

			itemExists, err := queries.ItemWithIdExists(db, nonexistentItemId)
			if assert.NoError(t, err) {
				if assert.False(t, itemExists) {
					payload := restapi.AddSalePayload{
						Items: []models.Id{},
					}
					url := path.Sales().String()
					request := test.CreatePostRequest(url, &payload)
					request.AddCookie(test.CreateCookie(sessionId))

					router.ServeHTTP(writer, request)

					if assert.Equal(t, http.StatusBadRequest, writer.Code) {
						sales, err := queries.GetSales(db)

						if assert.NoError(t, err) {
							assert.Empty(t, sales)
						}
					}
				}
			}
		})

		t.Run("Duplicate item", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			cashier := test.AddCashierToDatabase(db)
			seller := test.AddSellerToDatabase(db)
			item := test.AddItemToDatabase(db, seller.UserId, 1)
			sessionId := test.AddSessionToDatabase(db, cashier.UserId)

			payload := restapi.AddSalePayload{
				Items: []models.Id{item.ItemId, item.ItemId}, // Causes the operation to fail
			}
			url := path.Sales().String()
			request := test.CreatePostRequest(url, &payload)
			request.AddCookie(test.CreateCookie(sessionId))

			router.ServeHTTP(writer, request)

			if assert.Equal(t, http.StatusBadRequest, writer.Code) {
				sales, err := queries.GetSales(db)

				if assert.NoError(t, err) {
					assert.Empty(t, sales)
				}
			}
		})
	})
}
