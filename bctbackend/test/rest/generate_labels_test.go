//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"bctbackend/database/models"
	path "bctbackend/server/paths"
	"bctbackend/server/rest"
	restapi "bctbackend/server/rest"
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
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				seller, sessionId := setup.LoggedIn(setup.Seller())
				item1 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithFrozen(false), aux.WithHidden(false))

				url := path.Labels()
				request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
					Layout:  defaultLayout,
					ItemIds: []models.Id{item1.ItemID},
				}, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())
				setup.RequireFrozen(t, item1.ItemID)
			})

			t.Run("10 items", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				seller, sessionId := setup.LoggedIn(setup.Seller())

				items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
				itemIds := models.CollectItemIds(items)

				url := path.Labels()
				request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
					Layout:  defaultLayout,
					ItemIds: itemIds,
				}, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

				for _, item := range items {
					setup.RequireFrozen(t, item.ItemID)
				}
			})
		})

		t.Run("Multiple sellers", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())
			otherSeller := setup.Seller()

			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
			otherItems := setup.Items(otherSeller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
			itemIds := models.CollectItemIds(items)

			url := path.Labels()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: itemIds,
			}, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

			for _, item := range items {
				setup.RequireFrozen(t, item.ItemID)
			}

			for _, item := range otherItems {
				setup.RequireNotFrozen(t, item.ItemID)
			}
		})

		t.Run("Frozen items", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())

			items := setup.Items(seller.UserId, 10, aux.WithFrozen(true), aux.WithHidden(false))
			itemIds := models.CollectItemIds(items)

			url := path.Labels()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: itemIds,
			}, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

			for _, item := range items {
				setup.RequireFrozen(t, item.ItemID)
			}
		})

		t.Run("Duplicate items", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())

			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
			itemIds := models.CollectItemIds(items)

			url := path.Labels()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: append(itemIds, itemIds...),
			}, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

			for _, item := range items {
				setup.RequireFrozen(t, item.ItemID)
			}
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No items listed", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())

			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))

			url := path.Labels()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: []models.Id{},
			}, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "missing_items")

			for _, item := range items {
				setup.RequireNotFrozen(t, item.ItemID)
			}
		})

		t.Run("Nonexistent item", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())

			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
			nonexistendItemId := models.Id(1000)
			setup.RequireNoSuchItems(t, nonexistendItemId)

			url := path.Labels()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: []models.Id{nonexistendItemId},
			}, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_item")

			for _, item := range items {
				setup.RequireNotFrozen(t, item.ItemID)
			}
		})

		t.Run("As nonowning seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			owningSeller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Seller())

			items := setup.Items(owningSeller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))

			url := path.Labels()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: []models.Id{items[0].ItemID},
			}, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_seller")

			for _, item := range items {
				setup.RequireNotFrozen(t, item.ItemID)
			}
		})

		t.Run("As admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Admin())

			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))

			url := path.Labels()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: []models.Id{items[0].ItemID},
			}, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			for _, item := range items {
				setup.RequireNotFrozen(t, item.ItemID)
			}
		})

		t.Run("As cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Cashier())

			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))

			url := path.Labels()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: []models.Id{items[0].ItemID},
			}, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			for _, item := range items {
				setup.RequireNotFrozen(t, item.ItemID)
			}
		})

		t.Run("Without cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, _ := setup.LoggedIn(setup.Seller())

			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))

			url := path.Labels()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: []models.Id{items[0].ItemID},
			})
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")

			for _, item := range items {
				setup.RequireNotFrozen(t, item.ItemID)
			}
		})

		t.Run("Invalid session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, _ := setup.LoggedIn(setup.Seller())

			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))

			url := path.Labels()
			request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
				Layout:  defaultLayout,
				ItemIds: []models.Id{items[0].ItemID},
			}, WithSessionCookie("fake_session_id"))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")

			for _, item := range items {
				setup.RequireNotFrozen(t, item.ItemID)
			}
		})

		t.Run("Invalid layout", func(t *testing.T) {
			layouts := []rest.Layout{
				{
					PaperWidth:   0,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   -1,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  0,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  -1,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: -10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: -10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: -10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: -10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      0,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      -1,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         0,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         -1,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: -10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: -10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: -10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: -10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: -10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: -10, Left: 10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: -10, Right: 10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: -10},
					FontSize:     12,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     0,
				},
				{
					PaperWidth:   210,
					PaperHeight:  297,
					PaperMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					Columns:      2,
					Rows:         10,
					LabelMargins: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					LabelPadding: rest.Insets{Top: 10, Bottom: 10, Left: 10, Right: 10},
					FontSize:     -1,
				},
			}

			for _, layout := range layouts {
				testLabel := fmt.Sprintf("Layout %v", layout)
				t.Run(testLabel, func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					seller, sessionId := setup.LoggedIn(setup.Seller())
					items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
					itemIds := models.CollectItemIds(items)

					url := path.Labels()
					request := CreatePostRequest(url, &restapi.GenerateLabelsPayload{
						Layout:  layout,
						ItemIds: itemIds,
					}, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)
					RequireFailureType(t, writer, http.StatusForbidden, "invalid_layout")

					for _, item := range items {
						setup.RequireNotFrozen(t, item.ItemID)
					}
				})
			}
		})
	})
}
