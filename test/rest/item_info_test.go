//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"bctbackend/database/models"
	path "bctbackend/server/paths"
	restapi "bctbackend/server/rest"
	rest "bctbackend/server/shared"
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
					cashier, sessionID := setup.LoggedIn(setup.Cashier())
					item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))

					saleIDs := []models.ID{}
					for i := 0; i < sale_count; i++ {
						sale := setup.Sale(cashier.UserID, []models.ID{item.ItemID})
						saleIDs = append(saleIDs, sale.SaleID)
					}

					url := path.Item(item.ItemID)
					request := CreateGetRequest(url, WithSessionCookie(sessionID))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code)

					response := FromJson[restapi.GetItemInformationSuccessResponse](t, writer.Body.String())
					require.Equal(t, item.Description, response.Description)
					require.Equal(t, item.PriceInCents, response.PriceInCents)
					require.Equal(t, item.CategoryID, response.CategoryID)
					require.Equal(t, item.SellerID, response.SellerID)
					require.Equal(t, item.ItemID, response.ItemID)
					require.Equal(t, rest.ConvertTimestampToDateTime(item.AddedAt), response.AddedAt)
					require.Equal(t, item.Donation, *response.Donation)
					require.Equal(t, item.Charity, *response.Charity)
					require.Equal(t, item.Frozen, *response.Frozen)
					require.NotNil(t, response.SoldIn)
					require.Equal(t, saleIDs, *response.SoldIn)
				})
			}
		})

		t.Run("As admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			_, sessionID := setup.LoggedIn(setup.Admin())

			item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))

			url := path.Item(item.ItemID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			response := FromJson[restapi.GetItemInformationSuccessResponse](t, writer.Body.String())
			require.Equal(t, item.Description, response.Description)
			require.Equal(t, item.PriceInCents, response.PriceInCents)
			require.Equal(t, item.CategoryID, response.CategoryID)
			require.Equal(t, item.SellerID, response.SellerID)
			require.Equal(t, item.ItemID, response.ItemID)
			require.Equal(t, rest.ConvertTimestampToDateTime(item.AddedAt), response.AddedAt)
			require.Equal(t, item.Donation, *response.Donation)
			require.Equal(t, item.Charity, *response.Charity)
			require.Equal(t, item.Frozen, *response.Frozen)
			require.NotNil(t, response.SoldIn)
			require.Equal(t, []models.ID{}, *response.SoldIn)
		})

		t.Run("As owning seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionID := setup.LoggedIn(setup.Seller())
			item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))

			url := path.Item(item.ItemID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			response := FromJson[restapi.GetItemInformationSuccessResponse](t, writer.Body.String())
			require.Equal(t, item.Description, response.Description)
			require.Equal(t, item.PriceInCents, response.PriceInCents)
			require.Equal(t, item.CategoryID, response.CategoryID)
			require.Equal(t, item.SellerID, response.SellerID)
			require.Equal(t, item.ItemID, response.ItemID)
			require.Equal(t, rest.ConvertTimestampToDateTime(item.AddedAt), response.AddedAt)
			require.Equal(t, item.Donation, *response.Donation)
			require.Equal(t, item.Charity, *response.Charity)
			require.Equal(t, item.Frozen, *response.Frozen)
			require.NotNil(t, response.SoldIn)
			require.Equal(t, []models.ID{}, *response.SoldIn)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Invalid item ID", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionID := setup.LoggedIn(setup.Cashier())

			url := path.ItemStr("abc")
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_item_id")
		})

		t.Run("As nonowner seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionID := setup.LoggedIn(setup.Seller())
			ownerSeller := setup.Seller()
			item := setup.Item(ownerSeller.UserID, aux.WithDummyData(1), aux.WithHidden(false))

			url := path.Item(item.ItemID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_seller")
		})

		t.Run("Item does not exist", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			// Log in as cashier
			_, sessionID := setup.LoggedIn(setup.Cashier())

			// Get ID for nonexisting item
			nonexistentItem := models.ID(1)
			setup.RequireNoSuchItems(t, nonexistentItem)

			// Attempt to get information for nonexistent item
			url := path.Item(nonexistentItem)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))

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
			item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))

			for i := 0; i < sale_count; i++ {
				setup.Sale(cashier.UserID, []models.ID{item.ItemID})
			}

			url := path.Item(item.ItemID)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Session expired", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionID := setup.LoggedIn(setup.Seller(), aux.WithExpiration(100))
			item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))

			setup.Clock.Advance(200)

			url := path.Item(item.ItemID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusUnauthorized, writer.Code)
		})
	})
}
