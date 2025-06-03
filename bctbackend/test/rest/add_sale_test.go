//go:build test

package rest

import (
	"net/http"
	"testing"

	"bctbackend/database/models"
	"bctbackend/database/queries"
	rest_api "bctbackend/rest"
	"bctbackend/rest/path"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestAddSale(t *testing.T) {
	url := path.Sales().String()

	t.Run("Success", func(t *testing.T) {
		setup, router, writer := NewRestFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier, sessionId := setup.LoggedIn(setup.Cashier())
		item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

		payload := rest_api.AddSalePayload{
			Items: []models.Id{item.ItemID},
		}
		request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
		router.ServeHTTP(writer, request)
		require.Equal(t, http.StatusCreated, writer.Code)

		response := FromJson[rest_api.AddSaleSuccessResponse](t, writer.Body.String())

		sale, err := queries.GetSaleWithId(setup.Db, response.SaleId)
		require.NoError(t, err)
		require.Equal(t, cashier.UserId, sale.CashierId)

		saleItems, err := queries.GetSaleItems(setup.Db, sale.SaleId)
		require.NoError(t, err)
		require.Len(t, saleItems, 1)
		require.Equal(t, item.ItemID, saleItems[0].ItemID)
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Adding sale as seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

			payload := rest_api.AddSalePayload{
				Items: []models.Id{item.ItemID},
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			sales := []*models.SaleSummary{}
			err := queries.GetSales(setup.Db, queries.CollectTo(&sales))
			require.NoError(t, err)
			require.Empty(t, sales)
		})

		t.Run("Adding sale as admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			payload := rest_api.AddSalePayload{
				Items: []models.Id{item.ItemID},
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			sales := []*models.SaleSummary{}
			err := queries.GetSales(setup.Db, queries.CollectTo(&sales))
			require.NoError(t, err)
			require.Empty(t, sales)
		})

		t.Run("No items in sale", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Cashier())
			payload := rest_api.AddSalePayload{
				Items: []models.Id{},
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusForbidden, writer.Code)

			sales := []*models.SaleSummary{}
			err := queries.GetSales(setup.Db, queries.CollectTo(&sales))
			require.NoError(t, err)
			require.Empty(t, sales)
		})

		t.Run("Nonexistent item in sale", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Cashier())
			nonexistentItemId := models.Id(1000)
			setup.RequireNoSuchItems(t, nonexistentItemId)

			payload := rest_api.AddSalePayload{
				Items: []models.Id{nonexistentItemId},
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_item")

			sales := []*models.SaleSummary{}
			err := queries.GetSales(setup.Db, queries.CollectTo(&sales))
			require.NoError(t, err)
			require.Empty(t, sales)
		})

		t.Run("Duplicate item in sale", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Cashier())
			seller := setup.Seller()
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

			payload := rest_api.AddSalePayload{
				Items: []models.Id{item.ItemID, item.ItemID},
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "duplicate_item_in_sale")

			sales := []*models.SaleSummary{}
			err := queries.GetSales(setup.Db, queries.CollectTo(&sales))
			require.NoError(t, err)
			require.Empty(t, sales)
		})

		t.Run("Without cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

			payload := rest_api.AddSalePayload{
				Items: []models.Id{item.ItemID},
			}
			request := CreatePostRequest(url, &payload)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Cookie with fake session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

			payload := rest_api.AddSalePayload{
				Items: []models.Id{item.ItemID},
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie("fake_session_id"))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
		})
	})
}
