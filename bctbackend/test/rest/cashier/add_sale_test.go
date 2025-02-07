//go:build test

package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bctbackend/database/models"
	"bctbackend/database/queries"
	rest_api "bctbackend/rest/cashier"
	"bctbackend/rest/path"

	"bctbackend/test"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestAddSaleItem(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, router := test.CreateRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		seller := AddSellerToDatabase(db)
		cashier := AddCashierToDatabase(db)
		item := AddItemToDatabase(db, seller.UserId, WithDummyData(1))
		sessionId := test.AddSessionToDatabase(db, cashier.UserId)
		payload := rest_api.AddSalePayload{
			Items: []models.Id{item.ItemId},
		}
		url := path.Sales().String()
		request := test.CreatePostRequest(url, &payload)

		request.AddCookie(test.CreateCookie(sessionId))
		router.ServeHTTP(writer, request)
		require.Equal(t, http.StatusCreated, writer.Code)

		response := test.FromJson[rest_api.AddSaleSuccessResponse](writer.Body.String())

		sale, err := queries.GetSaleWithId(db, response.SaleId)
		require.NoError(t, err)
		require.Equal(t, cashier.UserId, sale.CashierId)

		saleItems, err := queries.GetSaleItems(db, sale.SaleId)
		require.NoError(t, err)
		require.Len(t, saleItems, 1)
		require.Equal(t, item.ItemId, saleItems[0].ItemId)
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("As seller", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := AddSellerToDatabase(db)
			item := AddItemToDatabase(db, seller.UserId, WithDummyData(1))
			sessionId := test.AddSessionToDatabase(db, seller.UserId) // Causes the operation to fail
			payload := rest_api.AddSalePayload{
				Items: []models.Id{item.ItemId},
			}
			url := path.Sales().String()
			request := test.CreatePostRequest(url, &payload)

			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusForbidden, writer.Code)

			sales := []*models.SaleSummary{}
			err := queries.GetSales(db, queries.CollectTo(&sales))
			require.NoError(t, err)
			require.Empty(t, sales)
		})

		t.Run("As admin", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			admin := AddAdminToDatabase(db)
			seller := AddSellerToDatabase(db)
			item := AddItemToDatabase(db, seller.UserId, WithDummyData(1))
			sessionId := test.AddSessionToDatabase(db, admin.UserId) // Causes the operation to fail
			payload := rest_api.AddSalePayload{
				Items: []models.Id{item.ItemId},
			}
			url := path.Sales().String()
			request := test.CreatePostRequest(url, &payload)

			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusForbidden, writer.Code)

			sales := []*models.SaleSummary{}
			err := queries.GetSales(db, queries.CollectTo(&sales))
			require.NoError(t, err)
			require.Empty(t, sales)
		})

		t.Run("No items", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			cashier := AddCashierToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, cashier.UserId)
			payload := rest_api.AddSalePayload{
				Items: []models.Id{},
			}
			url := path.Sales().String()
			request := test.CreatePostRequest(url, &payload)

			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusBadRequest, writer.Code)

			sales := []*models.SaleSummary{}
			err := queries.GetSales(db, queries.CollectTo(&sales))
			require.NoError(t, err)
			require.Empty(t, sales)
		})

		t.Run("Nonexistent item", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			cashier := AddCashierToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, cashier.UserId)
			nonexistentItemId := models.Id(1000)

			itemExists, err := queries.ItemWithIdExists(db, nonexistentItemId)
			require.NoError(t, err)
			require.False(t, itemExists)

			payload := rest_api.AddSalePayload{
				Items: []models.Id{},
			}
			url := path.Sales().String()
			request := test.CreatePostRequest(url, &payload)

			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusBadRequest, writer.Code)

			sales := []*models.SaleSummary{}
			err = queries.GetSales(db, queries.CollectTo(&sales))
			require.NoError(t, err)
			require.Empty(t, sales)
		})

		t.Run("Duplicate item", func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			cashier := AddCashierToDatabase(db)
			seller := AddSellerToDatabase(db)
			item := AddItemToDatabase(db, seller.UserId, WithDummyData(1))
			sessionId := test.AddSessionToDatabase(db, cashier.UserId)

			payload := rest_api.AddSalePayload{
				Items: []models.Id{item.ItemId, item.ItemId}, // Causes the operation to fail
			}
			url := path.Sales().String()
			request := test.CreatePostRequest(url, &payload)

			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusBadRequest, writer.Code)

			sales := []*models.SaleSummary{}
			err := queries.GetSales(db, queries.CollectTo(&sales))
			require.NoError(t, err)
			require.Empty(t, sales)
		})
	})
}
