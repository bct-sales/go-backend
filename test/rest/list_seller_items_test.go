//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	models "bctbackend/database/models"
	path "bctbackend/server/paths"
	"bctbackend/server/rest"
	shared "bctbackend/server/shared"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestListSellerItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("View own items", func(t *testing.T) {
			for _, sellerID := range []models.ID{models.ID(1), models.ID(2), models.ID(100)} {
				for _, itemCount := range []int{0, 1, 5, 100} {
					testLabel := fmt.Sprintf("SellerID: %d, ItemCount: %d", sellerID, itemCount)

					t.Run(testLabel, func(t *testing.T) {
						setup, router, writer := NewRestFixture(t, WithDefaultCategories)
						defer setup.Close()

						seller, sessionID := setup.LoggedIn(setup.Seller(aux.WithUserID(sellerID)))

						expectedItems := []*rest.GetSellerItemsItemData{}
						for i := 0; i < itemCount; i++ {
							item := setup.Item(seller.UserID, aux.WithDummyData(i), aux.WithHidden(false))
							expectedItems = append(expectedItems, &rest.GetSellerItemsItemData{
								ItemID:       item.ItemID,
								Description:  item.Description,
								PriceInCents: item.PriceInCents,
								CategoryID:   item.CategoryID,
								SellerID:     item.SellerID,
								AddedAt:      shared.ConvertTimestampToDateTime(item.AddedAt),
								Donation:     item.Donation,
								Charity:      item.Charity,
								Frozen:       item.Frozen,
							})
						}

						url := path.SellerItems(seller.UserID)
						request := CreateGetRequest(url, WithSessionCookie(sessionID))
						router.ServeHTTP(writer, request)
						require.Equal(t, http.StatusOK, writer.Code)

						actual := FromJSON[rest.GetSellerItemsSuccessResponse](t, writer.Body.String())
						require.NotNil(t, actual)
						require.Equal(t, expectedItems, actual.Items)
					})
				}
			}
		})

		t.Run("As admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			_, sessionID := setup.LoggedIn(setup.Admin())
			itemCount := 10

			expectedItems := []*rest.GetSellerItemsItemData{}
			for i := 0; i < itemCount; i++ {
				item := setup.Item(seller.UserID, aux.WithDummyData(i), aux.WithHidden(false))
				expectedItems = append(expectedItems, &rest.GetSellerItemsItemData{
					ItemID:       item.ItemID,
					Description:  item.Description,
					PriceInCents: item.PriceInCents,
					CategoryID:   item.CategoryID,
					SellerID:     item.SellerID,
					AddedAt:      shared.ConvertTimestampToDateTime(item.AddedAt),
					Donation:     item.Donation,
					Charity:      item.Charity,
					Frozen:       item.Frozen,
				})
			}

			url := path.SellerItems(seller.UserID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

			actual := FromJSON[rest.GetSellerItemsSuccessResponse](t, writer.Body.String())
			require.Equal(t, expectedItems, actual.Items)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Not logged in", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			itemCount := 10

			for i := 0; i < itemCount; i++ {
				setup.Item(seller.UserID, aux.WithDummyData(i), aux.WithHidden(false))
			}

			url := path.SellerItems(seller.UserID)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Seller accessing other seller's items", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			itemOwningSeller := setup.Seller()
			_, sessionID := setup.LoggedIn(setup.Seller())
			itemCount := 10

			for i := 0; i < itemCount; i++ {
				setup.Item(itemOwningSeller.UserID, aux.WithDummyData(i), aux.WithHidden(false))
			}

			url := path.SellerItems(itemOwningSeller.UserID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_seller")
		})

		t.Run("As cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			itemOwningSeller := setup.Seller()
			_, sessionID := setup.LoggedIn(setup.Cashier())
			itemCount := 10

			for i := 0; i < itemCount; i++ {
				setup.Item(itemOwningSeller.UserID, aux.WithDummyData(i), aux.WithHidden(false))
			}

			url := path.SellerItems(itemOwningSeller.UserID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
		})

		t.Run("Invalid seller id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller, sessionID := setup.LoggedIn(setup.Seller())
			itemCount := 10

			for i := 0; i < itemCount; i++ {
				setup.Item(seller.UserID, aux.WithDummyData(i), aux.WithHidden(false))
			}

			url := path.SellerItemsStr("xxx")
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_user_id")
		})

		t.Run("Listing items of nonexisting seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			_, sessionID := setup.LoggedIn(setup.Seller())
			nonexistentSellerID := setup.GenerateNonexistentUserID(t)

			url := path.SellerItems(nonexistentSellerID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_user")
		})

		t.Run("Listing items of nonseller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin()
			_, sessionID := setup.LoggedIn(setup.Seller())

			url := path.SellerItems(admin.UserID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_user")
		})

		t.Run("Without cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin()
			setup.LoggedIn(setup.Seller())

			url := path.SellerItems(admin.UserID)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Cookie with dummy session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin()
			setup.LoggedIn(setup.Seller())

			url := path.SellerItems(admin.UserID)
			request := CreateGetRequest(url, WithSessionCookie("dummy_session_id"))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
		})
	})
}
