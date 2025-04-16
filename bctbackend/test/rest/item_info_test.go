//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"bctbackend/database/models"
	"bctbackend/database/queries"
	restapi "bctbackend/rest/cashier"
	"bctbackend/rest/path"
	. "bctbackend/test"
	aux "bctbackend/test/helpers"

	"github.com/stretchr/testify/require"
)

func TestGetItemInformation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, sale_count := range []int{0, 1, 2, 5} {
			label := fmt.Sprintf("Sale count: %d", sale_count)

			t.Run(label, func(t *testing.T) {
				setup, router, writer := SetupRestTest()
				defer setup.Close()
				sale_count := 0

				seller := setup.Seller()
				cashier, sessionId := setup.LoggedIn(setup.Cashier())
				item := setup.Item(seller.UserId, aux.WithDummyData(1))

				for i := 0; i < sale_count; i++ {
					setup.Sale(cashier.UserId, []models.Id{item.ItemId})
				}

				url := path.SalesItems().WithItemId(item.ItemId)
				request := CreateGetRequest(url, WithCookie(sessionId))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				response := FromJson[restapi.GetItemInformationSuccessResponse](writer.Body.String())
				expectedHasBeenSold := sale_count > 0
				require.Equal(t, item.Description, response.Description)
				require.Equal(t, item.PriceInCents, response.PriceInCents)
				require.Equal(t, item.CategoryId, response.CategoryId)
				require.Equal(t, expectedHasBeenSold, *response.HasBeenSold)
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Invalid item ID", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Cashier())

			url := path.SalesItems().WithRawItemId("abc")
			request := CreateGetRequest(url, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_item_id")
		})

		t.Run("Wrong Role", func(t *testing.T) {
			t.Run("As seller", func(t *testing.T) {
				setup, router, writer := SetupRestTest()
				defer setup.Close()

				seller, sessionId := setup.LoggedIn(setup.Seller())
				item := setup.Item(seller.UserId, aux.WithDummyData(1))

				url := path.SalesItems().WithItemId(item.ItemId)
				request := CreateGetRequest(url, WithCookie(sessionId))
				router.ServeHTTP(writer, request)
				RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
			})

			t.Run("As admin", func(t *testing.T) {
				setup, router, writer := SetupRestTest()
				defer setup.Close()

				_, sessionId := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				item := setup.Item(seller.UserId, aux.WithDummyData(1))

				url := path.SalesItems().WithItemId(item.ItemId)
				request := CreateGetRequest(url, WithCookie(sessionId))
				router.ServeHTTP(writer, request)
				RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
			})
		})

		t.Run("Item does not exist", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()

			// Log in as cashier
			_, sessionId := setup.LoggedIn(setup.Cashier())

			// Get ID for nonexisting item
			nonexistentItem := models.NewId(1)

			// Sanity check: make sure item does not exist
			itemExists, err := queries.ItemWithIdExists(setup.Db, nonexistentItem)
			require.NoError(t, err)
			require.False(t, itemExists)

			// Attempt to get information for nonexistent item
			url := path.SalesItems().WithItemId(nonexistentItem)
			request := CreateGetRequest(url, WithCookie(sessionId))

			// Send request
			router.ServeHTTP(writer, request)

			// Check response
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_item")
		})

		t.Run("Not logged in", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()
			sale_count := 0

			seller := setup.Seller()
			cashier := setup.Cashier()
			item := setup.Item(seller.UserId, aux.WithDummyData(1))

			for i := 0; i < sale_count; i++ {
				setup.Sale(cashier.UserId, []models.Id{item.ItemId})
			}

			url := path.SalesItems().WithItemId(item.ItemId)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})
	})
}
