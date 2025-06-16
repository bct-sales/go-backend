//go:build test

package rest

import (
	"fmt"
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

func TestGetItemInformation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("As cashier", func(t *testing.T) {
			for _, sale_count := range []int{0, 1, 2, 5} {
				label := fmt.Sprintf("Sale count: %d", sale_count)

				t.Run(label, func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()
					sale_count := 0

					seller := setup.Seller()
					cashier, sessionId := setup.LoggedIn(setup.Cashier())
					item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

					saleIds := []models.Id{}
					for i := 0; i < sale_count; i++ {
						saleId := setup.Sale(cashier.UserId, []models.Id{item.ItemID})
						saleIds = append(saleIds, saleId)
					}

					url := path.Items().Id(item.ItemID)
					request := CreateGetRequest(url, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code)

					response := FromJson[restapi.GetItemInformationSuccessResponse](t, writer.Body.String())
					require.Equal(t, item.Description, response.Description)
					require.Equal(t, item.PriceInCents, response.PriceInCents)
					require.Equal(t, item.CategoryID, response.CategoryId)
					require.Equal(t, item.SellerID, response.SellerId)
					require.Equal(t, item.ItemID, response.ItemId)
					require.Equal(t, rest.ConvertTimestampToDateTime(item.AddedAt), response.AddedAt)
					require.Equal(t, item.Donation, *response.Donation)
					require.Equal(t, item.Charity, *response.Charity)
					require.Equal(t, item.Frozen, *response.Frozen)
					require.NotNil(t, response.SoldIn)
					require.Equal(t, saleIds, *response.SoldIn)
				})
			}
		})

		t.Run("As admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Admin())

			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

			url := path.Items().Id(item.ItemID)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			response := FromJson[restapi.GetItemInformationSuccessResponse](t, writer.Body.String())
			require.Equal(t, item.Description, response.Description)
			require.Equal(t, item.PriceInCents, response.PriceInCents)
			require.Equal(t, item.CategoryID, response.CategoryId)
			require.Equal(t, item.SellerID, response.SellerId)
			require.Equal(t, item.ItemID, response.ItemId)
			require.Equal(t, rest.ConvertTimestampToDateTime(item.AddedAt), response.AddedAt)
			require.Equal(t, item.Donation, *response.Donation)
			require.Equal(t, item.Charity, *response.Charity)
			require.Equal(t, item.Frozen, *response.Frozen)
			require.NotNil(t, response.SoldIn)
			require.Equal(t, []models.Id{}, *response.SoldIn)
		})

		t.Run("As owning seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

			url := path.Items().Id(item.ItemID)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			response := FromJson[restapi.GetItemInformationSuccessResponse](t, writer.Body.String())
			require.Equal(t, item.Description, response.Description)
			require.Equal(t, item.PriceInCents, response.PriceInCents)
			require.Equal(t, item.CategoryID, response.CategoryId)
			require.Equal(t, item.SellerID, response.SellerId)
			require.Equal(t, item.ItemID, response.ItemId)
			require.Equal(t, rest.ConvertTimestampToDateTime(item.AddedAt), response.AddedAt)
			require.Equal(t, item.Donation, *response.Donation)
			require.Equal(t, item.Charity, *response.Charity)
			require.Equal(t, item.Frozen, *response.Frozen)
			require.NotNil(t, response.SoldIn)
			require.Equal(t, []models.Id{}, *response.SoldIn)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Invalid item ID", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Cashier())

			url := path.Items().WithRawItemId("abc")
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_item_id")
		})

		t.Run("As nonowner seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Seller())
			ownerSeller := setup.Seller()
			item := setup.Item(ownerSeller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

			url := path.Items().Id(item.ItemID)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_seller")
		})

		t.Run("Item does not exist", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			// Log in as cashier
			_, sessionId := setup.LoggedIn(setup.Cashier())

			// Get ID for nonexisting item
			nonexistentItem := models.Id(1)
			setup.RequireNoSuchItems(t, nonexistentItem)

			// Attempt to get information for nonexistent item
			url := path.Items().Id(nonexistentItem)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))

			// Send request
			router.ServeHTTP(writer, request)

			// Check response
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_item")
		})

		t.Run("Not logged in", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()
			sale_count := 0

			seller := setup.Seller()
			cashier := setup.Cashier()
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

			for i := 0; i < sale_count; i++ {
				setup.Sale(cashier.UserId, []models.Id{item.ItemID})
			}

			url := path.Items().Id(item.ItemID)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})
	})
}
