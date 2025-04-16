//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/rest/path"
	. "bctbackend/test"
	aux "bctbackend/test/helpers"

	"github.com/stretchr/testify/require"
)

func TestListSellerItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, sellerId := range []models.Id{models.NewId(1), models.NewId(2), models.NewId(100)} {
			for _, itemCount := range []int{0, 1, 5, 100} {
				testLabel := fmt.Sprintf("SellerId: %d, ItemCount: %d", sellerId, itemCount)

				t.Run(testLabel, func(t *testing.T) {
					setup, router, writer := SetupRestTest()
					defer setup.Close()

					seller, sessionId := setup.LoggedIn(setup.Seller(aux.WithUserId(sellerId)))

					expectedItems := []models.Item{}
					for i := 0; i < itemCount; i++ {
						item := setup.Item(seller.UserId, aux.WithDummyData(i))
						expectedItems = append(expectedItems, *item)
					}

					url := path.SellerItems().WithSellerId(seller.UserId)
					request := CreateGetRequest(url, WithCookie(sessionId))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code)

					actual := FromJson[[]models.Item](writer.Body.String())
					require.Equal(t, expectedItems, *actual)
				})
			}
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Not logged in", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()

			seller := setup.Seller()
			itemCount := 10

			for i := 0; i < itemCount; i++ {
				setup.Item(seller.UserId, aux.WithDummyData(i))
			}

			url := path.SellerItems().WithSellerId(seller.UserId)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Wrong seller", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()

			itemOwningSeller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Seller())
			itemCount := 10

			for i := 0; i < itemCount; i++ {
				setup.Item(itemOwningSeller.UserId, aux.WithDummyData(i))
			}

			url := path.SellerItems().WithSellerId(itemOwningSeller.UserId)
			request := CreateGetRequest(url, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_seller")
		})

		t.Run("As cashier", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()

			itemOwningSeller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Cashier())
			itemCount := 10

			for i := 0; i < itemCount; i++ {
				setup.Item(itemOwningSeller.UserId, aux.WithDummyData(i))
			}

			url := path.SellerItems().WithSellerId(itemOwningSeller.UserId)
			request := CreateGetRequest(url, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
		})
	})
}
