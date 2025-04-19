//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"bctbackend/defs"
	restapi "bctbackend/rest"
	"bctbackend/rest/path"
	. "bctbackend/test/setup"

	models "bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"

	"github.com/stretchr/testify/require"
)

func TestAddSellerItem(t *testing.T) {
	t.Run("Successful", func(t *testing.T) {
		for _, sellerId := range []models.Id{models.NewId(1), models.NewId(2), models.NewId(100)} {
			for _, price := range []models.MoneyInCents{1, 100, 10000} {
				for _, description := range []string{"Xyz", "Test Description"} {
					for _, categoryId := range defs.ListCategories() {
						for _, donation := range []bool{true, false} {
							for _, charity := range []bool{true, false} {
								t.Run(fmt.Sprintf("sellerId=%d price=%d description=%s categoryId=%d donation=%t charity=%t", sellerId, price, description, categoryId, donation, charity), func(t *testing.T) {
									setup, router, writer := NewRestFixture()
									defer setup.Close()

									seller, sessionId := setup.LoggedIn(setup.Seller(aux.WithUserId(sellerId)))

									url := path.SellerItems().WithSellerId(seller.UserId)
									payload := restapi.AddSellerItemPayload{
										Price:       &price,
										Description: &description,
										CategoryId:  categoryId,
										Donation:    &donation,
										Charity:     &charity,
									}
									request := CreatePostRequest(url, &payload, WithCookie(sessionId))
									router.ServeHTTP(writer, request)

									require.Equal(t, http.StatusCreated, writer.Code)

									response := FromJson[restapi.AddSellerItemResponse](writer.Body.String())

									itemsInDatabase := []*models.Item{}
									err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase))
									require.NoError(t, err)
									require.Equal(t, 1, len(itemsInDatabase))

									itemInDatabase := itemsInDatabase[0]
									require.Equal(t, response.ItemId, itemInDatabase.ItemId)
									require.Equal(t, seller.UserId, itemInDatabase.SellerId)
									require.Equal(t, price, itemInDatabase.PriceInCents)
									require.Equal(t, description, itemInDatabase.Description)
									require.Equal(t, categoryId, itemInDatabase.CategoryId)
									require.Equal(t, donation, itemInDatabase.Donation)
									require.Equal(t, charity, itemInDatabase.Charity)
								})
							}
						}
					}
				}
			}
		}
	})

	t.Run("Failing", func(t *testing.T) {
		t.Run("Zero price", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryId := defs.Clothing50_56
			donation := false
			charity := false

			seller, sessionId := setup.LoggedIn(setup.Seller())

			url := path.SellerItems().WithSellerId(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "invalid_price")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase))
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Empty description", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := ""
			categoryId := defs.Shoes
			donation := false
			charity := false

			seller, sessionId := setup.LoggedIn(setup.Seller())

			url := path.SellerItems().WithSellerId(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "invalid_description")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase))
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid category", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			require.NotContains(t, defs.ListCategories(), categoryId)

			seller, sessionId := setup.LoggedIn(setup.Seller())

			url := path.SellerItems().WithSellerId(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_category")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase))
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding seller item as admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := defs.BabyChildEquipment
			donation := false
			charity := false

			seller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Admin())

			url := path.SellerItems().WithSellerId(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase))
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding seller item as cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := defs.Clothing104_116
			donation := false
			charity := false

			seller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Cashier())
			url := path.SellerItems().WithSellerId(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase))
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid url", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := defs.BabyChildEquipment
			donation := false
			charity := false

			_, sessionId := setup.LoggedIn(setup.Seller())

			url := path.SellerItems().WithRawSellerId("a")
			payload := restapi.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_user_id")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase))
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding as different seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := defs.BabyChildEquipment
			donation := false
			charity := false

			seller1 := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Seller())

			url := path.SellerItems().WithSellerId(seller1.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_user")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase))
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding item to nonexistent seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := defs.BabyChildEquipment
			donation := false
			charity := false

			_, sessionId := setup.LoggedIn(setup.Seller())
			nonexistentId := models.NewId(1000)

			userExists, err := queries.UserWithIdExists(setup.Db, nonexistentId)
			require.NoError(t, err)
			require.False(t, userExists)

			url := path.SellerItems().WithSellerId(nonexistentId)
			payload := restapi.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_user")

			itemsInDatabase := []*models.Item{}
			err = queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase))
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("No session ID in cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryId := defs.Clothing50_56
			donation := false
			charity := false

			seller := setup.Seller()

			url := path.SellerItems().WithSellerId(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase))
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid ID in cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryId := defs.Clothing50_56
			donation := false
			charity := false

			seller := setup.Seller()
			invalidSessionId := "xxx"

			url := path.SellerItems().WithSellerId(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithCookie(invalidSessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase))
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})
	})
}
