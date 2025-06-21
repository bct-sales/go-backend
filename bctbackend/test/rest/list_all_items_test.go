//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/rest"
	"bctbackend/rest/path"

	shared "bctbackend/rest/shared"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func FromModel(item *models.Item) *rest.GetItemsItemData {
	return &rest.GetItemsItemData{
		ItemId:       item.ItemID,
		AddedAt:      shared.ConvertTimestampToDateTime(item.AddedAt),
		Description:  item.Description,
		PriceInCents: item.PriceInCents,
		CategoryId:   item.CategoryID,
		SellerId:     item.SellerID,
		Donation:     item.Donation,
		Charity:      item.Charity,
		Frozen:       item.Frozen,
	}
}

type FailureResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func TestGetAllItems(t *testing.T) {
	url := path.Items().String()

	t.Run("Success", func(t *testing.T) {
		t.Run("No items", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())

			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			expected := rest.GetItemsSuccessResponse{
				Items:          []rest.GetItemsItemData{},
				TotalItemCount: 0,
			}
			actual := FromJson[rest.GetItemsSuccessResponse](t, writer.Body.String())
			require.Equal(t, expected, *actual)
		})

		t.Run("One item", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()

			addedAtTimestamp := models.Timestamp(100)
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithAddedAt(addedAtTimestamp), aux.WithHidden(false))

			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			expected := rest.GetItemsSuccessResponse{
				Items:          []rest.GetItemsItemData{*FromModel(item)},
				TotalItemCount: 1,
			}
			actual := FromJson[rest.GetItemsSuccessResponse](t, writer.Body.String())
			require.Equal(t, expected, *actual)
		})

		t.Run("Two items", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()
			addedAtTimestamp := models.Timestamp(500)
			item1 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithAddedAt(addedAtTimestamp), aux.WithHidden(false))
			item2 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithAddedAt(addedAtTimestamp), aux.WithHidden(false))

			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)

			require.Equal(t, http.StatusOK, writer.Code)

			expected := rest.GetItemsSuccessResponse{
				Items:          []rest.GetItemsItemData{*FromModel(item1), *FromModel(item2)},
				TotalItemCount: 2,
			}
			actual := FromJson[rest.GetItemsSuccessResponse](t, writer.Body.String())
			require.Equal(t, expected, *actual)
		})

		for _, limit := range []int{1, 2, 10} {
			testLabel := fmt.Sprintf("Limit %d", limit)
			t.Run(testLabel, func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				itemCount := 100

				_, sessionId := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				items := setup.Items(seller.UserId, itemCount, aux.WithHidden(false))

				url := path.Items().WithRowSelection(nil, &limit)
				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)

				require.Equal(t, http.StatusOK, writer.Code)

				expectedItems := items[:limit]
				response := FromJson[rest.GetItemsSuccessResponse](t, writer.Body.String())
				actualItems := response.Items
				require.Len(t, actualItems, limit)
				require.Equal(t, itemCount, response.TotalItemCount)

				for i := range limit {
					require.Equal(t, expectedItems[i].ItemID, actualItems[i].ItemId)
				}
			})
		}

		for _, offset := range []int{0, 1, 2, 10} {
			testLabel := fmt.Sprintf("Offset %d", offset)
			t.Run(testLabel, func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				itemCount := 100

				_, sessionId := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				items := setup.Items(seller.UserId, itemCount, aux.WithHidden(false))

				url := path.Items().WithRowSelection(&offset, nil)
				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)

				require.Equal(t, http.StatusOK, writer.Code)

				expectedItems := items[offset:]
				response := FromJson[rest.GetItemsSuccessResponse](t, writer.Body.String())
				actualItems := response.Items
				require.Len(t, actualItems, len(expectedItems))
				require.Equal(t, itemCount, response.TotalItemCount)

				for i := range len(expectedItems) - offset {
					require.Equal(t, expectedItems[i].ItemID, actualItems[i].ItemId)
				}
			})
		}

		for _, limit := range []int{1, 2, 10, 25} {
			for _, offset := range []int{0, 1, 2, 10, 25} {
				testLabel := fmt.Sprintf("Offset %d", offset)
				t.Run(testLabel, func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					itemCount := 100

					_, sessionId := setup.LoggedIn(setup.Admin())
					seller := setup.Seller()
					items := setup.Items(seller.UserId, itemCount, aux.WithHidden(false))

					url := path.Items().WithRowSelection(&offset, &limit)
					request := CreateGetRequest(url, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)

					require.Equal(t, http.StatusOK, writer.Code)

					expectedItems := items[offset : offset+limit]
					response := FromJson[rest.GetItemsSuccessResponse](t, writer.Body.String())
					actualItems := response.Items
					require.Len(t, actualItems, len(expectedItems))
					require.Equal(t, itemCount, response.TotalItemCount)

					for i := range len(expectedItems) - offset {
						require.Equal(t, expectedItems[i].ItemID, actualItems[i].ItemId)
					}
				})
			}
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Wrong role", func(t *testing.T) {
			for _, roleId := range []models.RoleId{models.NewSellerRoleId(), models.NewCashierRoleId()} {
				roleString := roleId.Name()

				t.Run("As "+roleString, func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					_, sessionId := setup.LoggedIn(setup.User(roleId))

					request := CreateGetRequest(url, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)

					RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
				})
			}
		})

		t.Run("No cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Cookie with fake session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			request := CreateGetRequest(url, WithSessionCookie("fake_session_id"))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
		})

		t.Run("Cookie without session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			request := CreateGetRequest(url, WithCookie("whatever", "whatever"))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})
	})
}
