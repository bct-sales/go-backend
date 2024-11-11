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
	"bctbackend/test"

	"github.com/stretchr/testify/assert"
)

func TestListAllItems(t *testing.T) {
	t.Run("No items", func(t *testing.T) {
		db, router := test.CreateRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		admin := test.AddAdminToDatabase(db)
		sessionId := test.AddSessionToDatabase(db, admin.UserId)

		url := path.Items().String()
		request, err := http.NewRequest("GET", url, nil)
		request.AddCookie(test.CreateCookie(sessionId))

		if assert.NoError(t, err) {
			router.ServeHTTP(writer, request)

			if assert.Equal(t, http.StatusOK, writer.Code) {
				expected := []models.Item{}
				actual := test.FromJson[[]models.Item](writer.Body.String())
				assert.Equal(t, expected, *actual)
			}
		}
	})

	t.Run("One item", func(t *testing.T) {
		db, router := test.CreateRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		adminId := test.AddAdminToDatabase(db).UserId
		sessionId := test.AddSessionToDatabase(db, adminId)

		sellerId := test.AddSellerToDatabase(db).UserId
		item := models.NewItem(0, 100, "test item", 1000, defs.Shoes, sellerId, false, false)
		itemId, err := queries.AddItem(db, item.Timestamp, item.Description, item.PriceInCents, item.CategoryId, item.SellerId, item.Donation, item.Charity)

		if !assert.NoError(t, err) {
			return
		}

		item.ItemId = itemId

		url := path.Items().String()
		request, err := http.NewRequest("GET", url, nil)
		request.AddCookie(test.CreateCookie(sessionId))

		if assert.NoError(t, err) {
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusOK, writer.Code)

			expected := []models.Item{*item}
			actual := test.FromJson[[]models.Item](writer.Body.String())
			assert.Equal(t, expected, *actual)
		}
	})

	t.Run("Two items", func(t *testing.T) {
		db, router := test.CreateRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		adminId := test.AddAdminToDatabase(db).UserId
		sessionId := test.AddSessionToDatabase(db, adminId)
		sellerId := test.AddSellerToDatabase(db).UserId
		item1 := models.NewItem(0, 100, "test item", 1000, defs.Shoes, sellerId, false, false)
		item2 := models.NewItem(0, 100, "test item", 1000, defs.Shoes, sellerId, false, false)

		itemId, err := queries.AddItem(db, item1.Timestamp, item1.Description, item1.PriceInCents, item1.CategoryId, item1.SellerId, item1.Donation, item1.Charity)
		if !assert.NoError(t, err) {
			return
		}
		item1.ItemId = itemId

		itemId, err = queries.AddItem(db, item2.Timestamp, item2.Description, item2.PriceInCents, item2.CategoryId, item2.SellerId, item2.Donation, item2.Charity)
		if !assert.NoError(t, err) {
			return
		}
		item2.ItemId = itemId

		url := path.Items().String()
		request, err := http.NewRequest("GET", url, nil)
		request.AddCookie(test.CreateCookie(sessionId))

		if assert.NoError(t, err) {
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusOK, writer.Code)

			expected := []models.Item{*item1, *item2}
			actual := test.FromJson[[]models.Item](writer.Body.String())
			assert.Equal(t, expected, *actual)
		}
	})
}

func TestListAllItemsAsNonAdmin(t *testing.T) {
	for _, roleId := range []models.Id{models.SellerRoleId, models.CashierRoleId} {
		roleString, err := models.NameOfRole(roleId)

		if err != nil {
			panic(err)
		}

		t.Run("As "+roleString, func(t *testing.T) {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			userId := test.AddUserToDatabase(db, roleId).UserId
			sessionId := test.AddSessionToDatabase(db, userId)

			url := path.Items().String()
			request, err := http.NewRequest("GET", url, nil)
			request.AddCookie(test.CreateCookie(sessionId))

			if assert.NoError(t, err) {
				router.ServeHTTP(writer, request)

				assert.Equal(t, http.StatusForbidden, writer.Code)
			}
		})
	}
}
