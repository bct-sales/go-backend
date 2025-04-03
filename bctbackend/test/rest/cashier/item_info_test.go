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

	test "bctbackend/test/rest"
	"bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestGetItemInformation(t *testing.T) {
	for _, sale_count := range []int{0, 1, 2, 5} {
		label := fmt.Sprintf("Sale count: %d", sale_count)

		t.Run(label, func(t *testing.T) {
			sale_count := 0
			db, router := setup.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := setup.AddSellerToDatabase(db)
			cashier := setup.AddCashierToDatabase(db)
			item := setup.AddItemToDatabase(db, seller.UserId, setup.WithDummyData(1))

			for i := 0; i < sale_count; i++ {
				setup.AddSaleToDatabase(db, cashier.UserId, []models.Id{item.ItemId})
			}

			sessionId := setup.Session(db, cashier.UserId)

			url := path.SalesItems().WithItemId(item.ItemId)
			request := setup.CreateGetRequest(url)

			request.AddCookie(setup.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			response := setup.FromJson[restapi.GetItemInformationSuccessResponse](writer.Body.String())
			expectedHasBeenSold := sale_count > 0
			require.Equal(t, item.Description, response.Description)
			require.Equal(t, item.PriceInCents, response.PriceInCents)
			require.Equal(t, item.CategoryId, response.CategoryId)
			require.Equal(t, expectedHasBeenSold, *response.HasBeenSold)
		})
	}
}

func TestGetItemInformationWithInvalidId(t *testing.T) {
	db, router := setup.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	cashier := setup.AddCashierToDatabase(db)
	sessionId := setup.Session(db, cashier.UserId)

	url := path.SalesItems().WithRawItemId("abc")
	request := setup.CreateGetRequest(url)

	request.AddCookie(setup.CreateCookie(sessionId))
	router.ServeHTTP(writer, request)
	require.Equal(t, http.StatusBadRequest, writer.Code)
}

func TestGetItemInformationAsSeller(t *testing.T) {
	db, router := setup.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	seller := setup.AddSellerToDatabase(db)
	sessionId := setup.Session(db, seller.UserId)
	item := setup.AddItemToDatabase(db, seller.UserId, setup.WithDummyData(1))

	setup.AddItemToDatabase(db, seller.UserId, setup.WithDummyData(1))

	url := path.SalesItems().WithItemId(item.ItemId)
	request := setup.CreateGetRequest(url)

	request.AddCookie(setup.CreateCookie(sessionId))
	router.ServeHTTP(writer, request)
	require.Equal(t, http.StatusForbidden, writer.Code)
}

func TestGetItemInformationAsAdmin(t *testing.T) {
	db, router := setup.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	admin := setup.AddAdminToDatabase(db)
	seller := setup.AddSellerToDatabase(db)
	sessionId := setup.Session(db, admin.UserId)
	item := setup.AddItemToDatabase(db, seller.UserId, setup.WithDummyData(1))

	setup.AddItemToDatabase(db, seller.UserId, setup.WithDummyData(1))

	url := path.SalesItems().WithItemId(item.ItemId)
	request := setup.CreateGetRequest(url)

	request.AddCookie(setup.CreateCookie(sessionId))
	router.ServeHTTP(writer, request)
	require.Equal(t, http.StatusForbidden, writer.Code)
}

func TestGetItemInformationWithNonexistentItem(t *testing.T) {
	db, router := setup.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	// Create cashier
	cashier := setup.AddCashierToDatabase(db)

	// Get ID for nonexisting item
	nonexistentItem := models.NewId(1)

	// Sanity check: make sure item does not exist
	itemExists, err := queries.ItemWithIdExists(db, nonexistentItem)
	require.NoError(t, err)
	require.False(t, itemExists)

	// Attempt to get information for nonexistent item
	url := path.SalesItems().WithItemId(nonexistentItem)
	request := setup.CreateGetRequest(url)

	// Act as cashier
	sessionId := setup.Session(db, cashier.UserId)
	request.AddCookie(setup.CreateCookie(sessionId))

	// Send request
	router.ServeHTTP(writer, request)

	// Check response
	test.RequireFailureType(t, writer, http.StatusNotFound, "no_such_item")
}
