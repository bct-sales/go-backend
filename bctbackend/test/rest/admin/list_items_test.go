//go:build test

package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	"bctbackend/rest/path"
	rest "bctbackend/rest/shared"
	"bctbackend/test"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

type Item struct {
	ItemId       models.Id                `json:"itemId"`
	AddedAt      rest.StructuredTimestamp `json:"addedAt"`
	Description  string                   `json:"description"`
	PriceInCents models.MoneyInCents      `json:"priceInCents"`
	CategoryId   models.Id                `json:"categoryId"`
	SellerId     models.Id                `json:"sellerId"`
	Donation     bool                     `json:"donation"`
	Charity      bool                     `json:"charity"`
}

type SuccessResponse struct {
	Items []Item `json:"items"`
}

func TestListAllItems(t *testing.T) {
	t.Run("Success with no items", func(t *testing.T) {
		db, router := test.CreateRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		admin := AddAdminToDatabase(db)
		sessionId := test.AddSessionToDatabase(db, admin.UserId)

		url := path.Items().String()
		request := test.CreateGetRequest(url)
		request.AddCookie(test.CreateCookie(sessionId))

		router.ServeHTTP(writer, request)
		require.Equal(t, http.StatusOK, writer.Code)

		expected := SuccessResponse{Items: []Item{}}
		actual := test.FromJson[SuccessResponse](writer.Body.String())
		require.Equal(t, expected, *actual)
	})

	t.Run("Success with one item", func(t *testing.T) {
		db, router := test.CreateRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		adminId := AddAdminToDatabase(db).UserId
		sessionId := test.AddSessionToDatabase(db, adminId)

		sellerId := AddSellerToDatabase(db).UserId
		addedAtTimestamp := models.Timestamp(100)
		item := Item{
			ItemId:       0,
			AddedAt:      rest.FromTimestamp(addedAtTimestamp),
			Description:  "test item",
			PriceInCents: 1000,
			CategoryId:   defs.Shoes,
			SellerId:     sellerId,
			Donation:     false,
			Charity:      false,
		}
		itemId, err := queries.AddItem(db, models.Timestamp(addedAtTimestamp), item.Description, item.PriceInCents, item.CategoryId, item.SellerId, item.Donation, item.Charity)
		require.NoError(t, err)

		url := path.Items().String()
		request := test.CreateGetRequest(url)
		request.AddCookie(test.CreateCookie(sessionId))

		router.ServeHTTP(writer, request)
		require.Equal(t, http.StatusOK, writer.Code)

		item.ItemId = itemId
		expected := SuccessResponse{
			Items: []Item{item},
		}
		actual := test.FromJson[SuccessResponse](writer.Body.String())
		require.Equal(t, expected, *actual)
	})

	t.Run("Success with wo items", func(t *testing.T) {
		db, router := test.CreateRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		adminId := AddAdminToDatabase(db).UserId
		sessionId := test.AddSessionToDatabase(db, adminId)
		sellerId := AddSellerToDatabase(db).UserId
		addedAtTimestamp := models.Timestamp(500)
		item1 := Item{
			ItemId:       0,
			AddedAt:      rest.FromTimestamp(addedAtTimestamp),
			Description:  "test item",
			PriceInCents: 1000,
			CategoryId:   defs.Shoes,
			SellerId:     sellerId,
			Donation:     false,
			Charity:      false}
		item2 := Item{
			ItemId:       0,
			AddedAt:      rest.FromTimestamp(addedAtTimestamp),
			Description:  "test item 2",
			PriceInCents: 5000,
			CategoryId:   defs.Clothing128_140,
			SellerId:     sellerId,
			Donation:     false,
			Charity:      false}

		itemId1, err := queries.AddItem(db, addedAtTimestamp, item1.Description, item1.PriceInCents, item1.CategoryId, item1.SellerId, item1.Donation, item1.Charity)
		require.NoError(t, err)

		itemId2, err := queries.AddItem(db, addedAtTimestamp, item2.Description, item2.PriceInCents, item2.CategoryId, item2.SellerId, item2.Donation, item2.Charity)
		require.NoError(t, err)

		url := path.Items().String()
		request := test.CreateGetRequest(url)
		request.AddCookie(test.CreateCookie(sessionId))

		router.ServeHTTP(writer, request)

		require.Equal(t, http.StatusOK, writer.Code)

		item1.ItemId = itemId1
		item2.ItemId = itemId2
		expected := SuccessResponse{
			Items: []Item{item1, item2},
		}
		actual := test.FromJson[SuccessResponse](writer.Body.String())
		require.Equal(t, expected, *actual)
	})

	t.Run("Failure due no admin role", func(t *testing.T) {
		for _, roleId := range []models.Id{models.SellerRoleId, models.CashierRoleId} {
			roleString, err := models.NameOfRole(roleId)

			if err != nil {
				panic(err)
			}

			t.Run("As "+roleString, func(t *testing.T) {
				db, router := test.CreateRestRouter()
				writer := httptest.NewRecorder()
				defer db.Close()

				userId := AddUserToDatabase(db, roleId).UserId
				sessionId := test.AddSessionToDatabase(db, userId)

				url := path.Items().String()
				request := test.CreateGetRequest(url)
				request.AddCookie(test.CreateCookie(sessionId))
				router.ServeHTTP(writer, request)

				require.Equal(t, http.StatusForbidden, writer.Code)
			})
		}
	})
}
