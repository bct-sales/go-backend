//go:build test

package rest

import (
	"net/http"
	"testing"

	"bctbackend/algorithms"
	"bctbackend/database/models"
	"bctbackend/rest"
	restapi "bctbackend/rest"
	"bctbackend/rest/path"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestGenerateLabels(t *testing.T) {
	defaultLayout := rest.Layout{
		PaperWidth:   210,
		PaperHeight:  297,
		PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
		Columns:      2,
		Rows:         10,
		LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
		LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
		FontSize:     12,
	}

	t.Run("Success", func(t *testing.T) {
		t.Run("Single seller", func(t *testing.T) {
			t.Run("Single item", func(t *testing.T) {
				setup, router, writer := NewRestFixture()
				defer setup.Close()

				seller, sessionId := setup.LoggedIn(setup.Seller())
				item1 := setup.Item(seller.UserId, aux.WithDummyData(1))

				url := path.Labels().String()
				request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
					Layout:  defaultLayout,
					ItemIds: []models.Id{item1.ItemId},
				}, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())
				setup.RequireFrozen(t, item1.ItemId)
			})
		})

		t.Run("Single seller", func(t *testing.T) {
			t.Run("10 items", func(t *testing.T) {
				setup, router, writer := NewRestFixture()
				defer setup.Close()

				seller, sessionId := setup.LoggedIn(setup.Seller())

				items := setup.Items(seller.UserId, 10)
				itemIds := algorithms.Map(items, func(item *models.Item) models.Id { return item.ItemId })

				url := path.Labels().String()
				request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
					Layout:  defaultLayout,
					ItemIds: itemIds,
				}, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

				for _, item := range items {
					setup.RequireFrozen(t, item.ItemId)
				}
			})
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No items listed", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())

			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false))

			url := path.Labels().String()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: []models.Id{},
			}, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "missing_items")

			for _, item := range items {
				setup.RequireNotFrozen(t, item.ItemId)
			}
		})

		t.Run("Nonexistent item", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())

			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false))
			nonexistendItemId := models.Id(1000)
			setup.RequireNoSuchItem(t, nonexistendItemId)

			url := path.Labels().String()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: []models.Id{nonexistendItemId},
			}, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_item")

			for _, item := range items {
				setup.RequireNotFrozen(t, item.ItemId)
			}
		})

		t.Run("As nonowning seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			owningSeller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Seller())

			items := setup.Items(owningSeller.UserId, 10, aux.WithFrozen(false))

			url := path.Labels().String()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: []models.Id{items[0].ItemId},
			}, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_seller")

			for _, item := range items {
				setup.RequireNotFrozen(t, item.ItemId)
			}
		})
	})
}
