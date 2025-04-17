//go:build test

package rest

import (
	"net/http"
	"testing"

	"bctbackend/rest/path"
	. "bctbackend/test/setup"

	models "bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"

	"github.com/stretchr/testify/require"
)

func TestUpdateItem(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Updating description", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())
			originalDescription := "old description"
			newDescription := "new description"
			originalItem := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithDescription(originalDescription))

			url := path.Items().Id(originalItem.ItemId)
			payload := struct {
				Description string `json:"description"`
			}{
				Description: newDescription,
			}
			request := CreatePutRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusNoContent, writer.Code)

			actualItem, err := queries.GetItemWithId(setup.Db, originalItem.ItemId)
			require.NoError(t, err)

			expectedItem := *originalItem
			expectedItem.Description = newDescription
			require.Equal(t, expectedItem, *actualItem)
		})

		t.Run("Updating as admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			seller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Admin())
			originalDescription := "old description"
			newDescription := "new description"
			originalItem := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithDescription(originalDescription))

			url := path.Items().Id(originalItem.ItemId)
			payload := struct {
				Description string `json:"description"`
			}{
				Description: newDescription,
			}
			request := CreatePutRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusNoContent, writer.Code)

			actualItem, err := queries.GetItemWithId(setup.Db, originalItem.ItemId)
			require.NoError(t, err)

			expectedItem := *originalItem
			expectedItem.Description = newDescription
			require.Equal(t, expectedItem, *actualItem)
		})

		t.Run("Updating price", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())
			originalPrice := 100
			newPrice := 200
			originalItem := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithPriceInCents(models.MoneyInCents(originalPrice)))

			url := path.Items().Id(originalItem.ItemId)
			payload := struct {
				PriceInCents int `json:"priceInCents"`
			}{
				PriceInCents: newPrice,
			}
			request := CreatePutRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusNoContent, writer.Code)

			actualItem, err := queries.GetItemWithId(setup.Db, originalItem.ItemId)
			require.NoError(t, err)

			expectedItem := *originalItem
			expectedItem.PriceInCents = models.MoneyInCents(newPrice)
			require.Equal(t, expectedItem, *actualItem)
		})

		t.Run("Updating charity and donation", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())
			originalItem := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithCharity(false), aux.WithDonation(false))

			url := path.Items().Id(originalItem.ItemId)
			payload := struct {
				Donation bool `json:"donation"`
				Charity  bool `json:"charity"`
			}{
				Donation: true,
				Charity:  true,
			}
			request := CreatePutRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusNoContent, writer.Code)

			actualItem, err := queries.GetItemWithId(setup.Db, originalItem.ItemId)
			require.NoError(t, err)

			expectedItem := *originalItem
			expectedItem.Donation = true
			expectedItem.Charity = true
			require.Equal(t, expectedItem, *actualItem)
		})
	})

	t.Run("Failures", func(t *testing.T) {
		t.Run("Updating frozen item", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())
			originalItem := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithFrozen(true))

			url := path.Items().Id(originalItem.ItemId)
			payload := struct {
				Description string `json:"description"`
			}{
				Description: "updated",
			}
			request := CreatePutRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "item_frozen")
			require.Equal(t, http.StatusForbidden, writer.Code)

			actualItem, err := queries.GetItemWithId(setup.Db, originalItem.ItemId)
			require.NoError(t, err)

			expectedItem := *originalItem
			require.Equal(t, expectedItem, *actualItem)
		})

		t.Run("Invalid price", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())
			originalItem := setup.Item(seller.UserId, aux.WithDummyData(1))

			url := path.Items().Id(originalItem.ItemId)
			payload := struct {
				PriceInCents int `json:"priceInCents"`
			}{
				PriceInCents: -100,
			}
			request := CreatePutRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "invalid_price")

			actualItem, err := queries.GetItemWithId(setup.Db, originalItem.ItemId)
			require.NoError(t, err)

			expectedItem := *originalItem
			require.Equal(t, expectedItem, *actualItem)
		})

		t.Run("Nonexisting item", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Seller())

			nonexistingItemId := models.Id(123)
			url := path.Items().Id(nonexistingItemId)
			setup.RequireNoSuchItem(t, nonexistingItemId)

			payload := struct {
				PriceInCents int `json:"priceInCents"`
			}{
				PriceInCents: 100,
			}
			request := CreatePutRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_item")
		})

		t.Run("Updating as wrong seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			ownerSeller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Seller())
			originalItem := setup.Item(ownerSeller.UserId, aux.WithDummyData(1))

			url := path.Items().Id(originalItem.ItemId)
			payload := struct {
				PriceInCents int `json:"priceInCents"`
			}{
				PriceInCents: 100,
			}
			request := CreatePutRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_seller")

			actualItem, err := queries.GetItemWithId(setup.Db, originalItem.ItemId)
			require.NoError(t, err)

			expectedItem := *originalItem
			require.Equal(t, expectedItem, *actualItem)
		})

		t.Run("Updating as cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			ownerSeller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Cashier())
			originalItem := setup.Item(ownerSeller.UserId, aux.WithDummyData(1))

			url := path.Items().Id(originalItem.ItemId)
			payload := struct {
				PriceInCents int `json:"priceInCents"`
			}{
				PriceInCents: 100,
			}
			request := CreatePutRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			actualItem, err := queries.GetItemWithId(setup.Db, originalItem.ItemId)
			require.NoError(t, err)

			expectedItem := *originalItem
			require.Equal(t, expectedItem, *actualItem)
		})
	})
}
