//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/server/path"
	"bctbackend/server/rest"
	shared "bctbackend/server/shared"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestListSellerItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("View own items", func(t *testing.T) {
			for _, sellerId := range []models.Id{models.Id(1), models.Id(2), models.Id(100)} {
				for _, itemCount := range []int{0, 1, 5, 100} {
					testLabel := fmt.Sprintf("SellerId: %d, ItemCount: %d", sellerId, itemCount)

					t.Run(testLabel, func(t *testing.T) {
						setup, router, writer := NewRestFixture(WithDefaultCategories)
						defer setup.Close()

						seller, sessionId := setup.LoggedIn(setup.Seller(aux.WithUserId(sellerId)))

						expectedItems := []*rest.GetSellerItemsItemData{}
						for i := 0; i < itemCount; i++ {
							item := setup.Item(seller.UserId, aux.WithDummyData(i), aux.WithHidden(false))
							expectedItems = append(expectedItems, &rest.GetSellerItemsItemData{
								ItemId:       item.ItemID,
								Description:  item.Description,
								PriceInCents: item.PriceInCents,
								CategoryId:   item.CategoryID,
								SellerId:     item.SellerID,
								AddedAt:      shared.ConvertTimestampToDateTime(item.AddedAt),
								Donation:     item.Donation,
								Charity:      item.Charity,
								Frozen:       item.Frozen,
							})
						}

						url := path.SellerItems().WithSellerId(seller.UserId)
						request := CreateGetRequest(url, WithSessionCookie(sessionId))
						router.ServeHTTP(writer, request)
						require.Equal(t, http.StatusOK, writer.Code)

						actual := FromJson[rest.GetSellerItemsSuccessResponse](t, writer.Body.String())
						require.NotNil(t, actual)
						require.Equal(t, expectedItems, actual.Items)
					})
				}
			}
		})

		t.Run("As admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Admin())
			itemCount := 10

			expectedItems := []*rest.GetSellerItemsItemData{}
			for i := 0; i < itemCount; i++ {
				item := setup.Item(seller.UserId, aux.WithDummyData(i), aux.WithHidden(false))
				expectedItems = append(expectedItems, &rest.GetSellerItemsItemData{
					ItemId:       item.ItemID,
					Description:  item.Description,
					PriceInCents: item.PriceInCents,
					CategoryId:   item.CategoryID,
					SellerId:     item.SellerID,
					AddedAt:      shared.ConvertTimestampToDateTime(item.AddedAt),
					Donation:     item.Donation,
					Charity:      item.Charity,
					Frozen:       item.Frozen,
				})
			}

			url := path.SellerItems().WithSellerId(seller.UserId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

			actual := FromJson[rest.GetSellerItemsSuccessResponse](t, writer.Body.String())
			require.Equal(t, expectedItems, actual.Items)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Not logged in", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			itemCount := 10

			for i := 0; i < itemCount; i++ {
				setup.Item(seller.UserId, aux.WithDummyData(i), aux.WithHidden(false))
			}

			url := path.SellerItems().WithSellerId(seller.UserId)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Seller accessing other seller's items", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			itemOwningSeller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Seller())
			itemCount := 10

			for i := 0; i < itemCount; i++ {
				setup.Item(itemOwningSeller.UserId, aux.WithDummyData(i), aux.WithHidden(false))
			}

			url := path.SellerItems().WithSellerId(itemOwningSeller.UserId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_seller")
		})

		t.Run("As cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			itemOwningSeller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Cashier())
			itemCount := 10

			for i := 0; i < itemCount; i++ {
				setup.Item(itemOwningSeller.UserId, aux.WithDummyData(i), aux.WithHidden(false))
			}

			url := path.SellerItems().WithSellerId(itemOwningSeller.UserId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
		})

		t.Run("Invalid seller id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())
			itemCount := 10

			for i := 0; i < itemCount; i++ {
				setup.Item(seller.UserId, aux.WithDummyData(i), aux.WithHidden(false))
			}

			url := path.SellerItems().WithRawSellerId("xxx")
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_user_id")
		})

		t.Run("Listing items of nonexisting seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Seller())
			nonexistentSellerId := models.Id(1000)
			setup.RequireNoSuchUsers(t, nonexistentSellerId)

			url := path.SellerItems().WithSellerId(nonexistentSellerId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_user")
		})

		t.Run("Listing items of nonseller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin()
			_, sessionId := setup.LoggedIn(setup.Seller())

			url := path.SellerItems().WithSellerId(admin.UserId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_user")
		})

		t.Run("Without cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin()
			setup.LoggedIn(setup.Seller())

			url := path.SellerItems().WithSellerId(admin.UserId)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Cookie with dummy session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin()
			setup.LoggedIn(setup.Seller())

			url := path.SellerItems().WithSellerId(admin.UserId)
			request := CreateGetRequest(url, WithSessionCookie("dummy_session_id"))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
		})
	})
}
