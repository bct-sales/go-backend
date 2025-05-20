//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	restapi "bctbackend/rest"
	"bctbackend/rest/path"
	. "bctbackend/test/setup"

	models "bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"

	"github.com/stretchr/testify/require"
)

func TestAddSellerItem(t *testing.T) {
	defaultCategoryTable := DefaultCategoryTable()

	t.Run("Successful", func(t *testing.T) {
		for _, sellerId := range []models.Id{models.NewId(1), models.NewId(2), models.NewId(100)} {
			for _, price := range []models.MoneyInCents{1, 100, 10000} {
				for _, description := range []string{"Xyz", "Test Description"} {
					for categoryId, _ := range defaultCategoryTable {
						for _, donation := range []bool{true, false} {
							for _, charity := range []bool{true, false} {
								t.Run(fmt.Sprintf("sellerId=%d price=%d description=%s categoryId=%d donation=%t charity=%t", sellerId, price, description, categoryId, donation, charity), func(t *testing.T) {
									setup, router, writer := NewRestFixture(WithDefaultCategories)
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
									request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
									router.ServeHTTP(writer, request)

									require.Equal(t, http.StatusCreated, writer.Code)
									response := FromJson[restapi.AddSellerItemResponse](t, writer.Body.String())

									itemsInDatabase := []*models.Item{}
									err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems)
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
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryId := CategoryId_Clothing50_56
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
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "invalid_price")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Empty description", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := ""
			categoryId := CategoryId_Shoes
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
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "invalid_item_description")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid category", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			require.NotContains(t, defaultCategoryTable, categoryId)

			seller, sessionId := setup.LoggedIn(setup.Seller())

			url := path.SellerItems().WithSellerId(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_category")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding seller item as admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := CategoryId_BabyChildEquipment
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
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding seller item as cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := CategoryId_Clothing104_116
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
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid url", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := CategoryId_BabyChildEquipment
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
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_user_id")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding as different seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := CategoryId_BabyChildEquipment
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
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_seller")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding item to nonexistent seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := CategoryId_BabyChildEquipment
			donation := false
			charity := false

			_, sessionId := setup.LoggedIn(setup.Seller())
			nonexistentUserId := models.NewId(1000)
			setup.RequireNoSuchUser(t, nonexistentUserId)

			url := path.SellerItems().WithSellerId(nonexistentUserId)
			payload := restapi.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_user")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("No session ID in cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryId := CategoryId_Clothing50_56
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
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid session ID in cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryId := CategoryId_Clothing50_56
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
			request := CreatePostRequest(url, &payload, WithSessionCookie(invalidSessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})
	})
}
