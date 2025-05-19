//go:build test

package rest

import (
	"net/http"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/rest/path"
	rest "bctbackend/rest/shared"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

type Item struct {
	ItemId       models.Id           `json:"itemId"`
	AddedAt      rest.DateTime       `json:"addedAt"`
	Description  string              `json:"description"`
	PriceInCents models.MoneyInCents `json:"priceInCents"`
	CategoryId   models.Id           `json:"categoryId"`
	SellerId     models.Id           `json:"sellerId"`
	Donation     bool                `json:"donation"`
	Charity      bool                `json:"charity"`
	Frozen       bool                `json:"frozen"`
}

func FromModel(item *models.Item) *Item {
	return &Item{
		ItemId:       item.ItemId,
		AddedAt:      rest.ConvertTimestampToDateTime(item.AddedAt),
		Description:  item.Description,
		PriceInCents: item.PriceInCents,
		CategoryId:   item.CategoryId,
		SellerId:     item.SellerId,
		Donation:     item.Donation,
		Charity:      item.Charity,
		Frozen:       item.Frozen,
	}
}

type SuccessResponse struct {
	Items []Item `json:"items"`
}

type FailureResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func TestGetAllItems(t *testing.T) {
	url := path.Items().String()

	t.Run("Success", func(t *testing.T) {
		t.Run("No items", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())

			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			expected := SuccessResponse{Items: []Item{}}
			actual := FromJson[SuccessResponse](t, writer.Body.String())
			require.Equal(t, expected, *actual)
		})

		t.Run("One item", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()

			addedAtTimestamp := models.Timestamp(100)
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithAddedAt(addedAtTimestamp), aux.WithHidden(false))

			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			expected := SuccessResponse{
				Items: []Item{*FromModel(item)},
			}
			actual := FromJson[SuccessResponse](t, writer.Body.String())
			require.Equal(t, expected, *actual)
		})

		t.Run("Two items", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()
			addedAtTimestamp := models.Timestamp(500)
			item1 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithAddedAt(addedAtTimestamp), aux.WithHidden(false))
			item2 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithAddedAt(addedAtTimestamp), aux.WithHidden(false))

			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)

			require.Equal(t, http.StatusOK, writer.Code)

			expected := SuccessResponse{
				Items: []Item{*FromModel(item1), *FromModel(item2)},
			}
			actual := FromJson[SuccessResponse](t, writer.Body.String())
			require.Equal(t, expected, *actual)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Wrong role", func(t *testing.T) {
			for _, roleId := range []models.Id{models.SellerRoleId, models.CashierRoleId} {
				roleString, err := models.NameOfRole(roleId)

				if err != nil {
					panic(err)
				}

				t.Run("As "+roleString, func(t *testing.T) {
					setup, router, writer := NewRestFixture()
					defer setup.Close()

					_, sessionId := setup.LoggedIn(setup.User(roleId))

					request := CreateGetRequest(url, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)

					RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
				})
			}
		})

		t.Run("No cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Cookie with fake session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			request := CreateGetRequest(url, WithSessionCookie("fake_session_id"))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
		})

		t.Run("Cookie without session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			request := CreateGetRequest(url, WithCookie("whatever", "whatever"))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})
	})
}
