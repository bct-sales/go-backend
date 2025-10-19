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

func FromModel(item *models.Item) *rest.ListItemsItemData {
	return &rest.ListItemsItemData{
		ItemID:       item.ItemID,
		AddedAt:      shared.ConvertTimestampToDateTime(item.AddedAt),
		Description:  item.Description,
		PriceInCents: item.PriceInCents,
		CategoryID:   item.CategoryID,
		SellerID:     item.SellerID,
		Donation:     item.Donation,
		Charity:      item.Charity,
		Frozen:       item.Frozen,
	}
}

type FailureResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func TestListAllItems(t *testing.T) {
	url := path.Items()

	t.Run("Success", func(t *testing.T) {
		for _, loggedInRole := range []models.RoleID{models.NewAdminRoleID(), models.NewCashierRoleID()} {
			testLabel := fmt.Sprintf("Logged in as %s", loggedInRole.Name())
			t.Run(testLabel, func(t *testing.T) {
				t.Run("No items", func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					_, sessionID := setup.LoggedIn(setup.User(loggedInRole))

					request := CreateGetRequest(url, WithSessionCookie(sessionID))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code)

					expected := rest.ListItemsSuccessResponse{
						Items: []rest.ListItemsItemData{},
						ListItemsStatistics: rest.ListItemsStatistics{
							TotalItemCount: 0,
							TotalItemValue: models.MoneyInCents(0),
						},
					}
					actual := FromJSON[rest.ListItemsSuccessResponse](t, writer.Body.String())
					require.Equal(t, expected, *actual)
				})

				t.Run("One item", func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					_, sessionID := setup.LoggedIn(setup.User(loggedInRole))
					seller := setup.Seller()

					addedAtTimestamp := models.Timestamp(100)
					item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithAddedAt(addedAtTimestamp), aux.WithHidden(false))

					request := CreateGetRequest(url, WithSessionCookie(sessionID))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code)

					expected := rest.ListItemsSuccessResponse{
						Items: []rest.ListItemsItemData{*FromModel(item)},
						ListItemsStatistics: rest.ListItemsStatistics{
							TotalItemCount: 1,
							TotalItemValue: item.PriceInCents,
						},
					}
					actual := FromJSON[rest.ListItemsSuccessResponse](t, writer.Body.String())
					require.Equal(t, expected, *actual)
				})

				t.Run("Two items", func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					_, sessionID := setup.LoggedIn(setup.User(loggedInRole))
					seller := setup.Seller()
					addedAtTimestamp := models.Timestamp(500)
					item1 := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithAddedAt(addedAtTimestamp), aux.WithHidden(false))
					item2 := setup.Item(seller.UserID, aux.WithDummyData(2), aux.WithAddedAt(addedAtTimestamp), aux.WithHidden(false))

					request := CreateGetRequest(url, WithSessionCookie(sessionID))
					router.ServeHTTP(writer, request)

					require.Equal(t, http.StatusOK, writer.Code)

					expected := rest.ListItemsSuccessResponse{
						Items: []rest.ListItemsItemData{*FromModel(item1), *FromModel(item2)},
						ListItemsStatistics: rest.ListItemsStatistics{
							TotalItemCount: 2,
							TotalItemValue: item1.PriceInCents + item2.PriceInCents,
						},
					}
					actual := FromJSON[rest.ListItemsSuccessResponse](t, writer.Body.String())
					require.Equal(t, expected, *actual)
				})

				t.Run("Only visible items", func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					_, sessionID := setup.LoggedIn(setup.User(loggedInRole))
					seller := setup.Seller()
					item1 := setup.Item(seller.UserID, aux.WithPriceInCents(100), aux.WithHidden(false))
					item2 := setup.Item(seller.UserID, aux.WithPriceInCents(200), aux.WithHidden(false))
					setup.Item(seller.UserID, aux.WithPriceInCents(400), aux.WithHidden(true))

					url := path.Items().AddHidden(false)
					request := CreateGetRequest(url, WithSessionCookie(sessionID))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code, writer.Body)

					expected := rest.ListItemsSuccessResponse{
						Items: []rest.ListItemsItemData{*FromModel(item1), *FromModel(item2)},
						ListItemsStatistics: rest.ListItemsStatistics{
							TotalItemCount: 2,
							TotalItemValue: item1.PriceInCents + item2.PriceInCents,
						},
					}
					actual := FromJSON[rest.ListItemsSuccessResponse](t, writer.Body.String())
					require.Equal(t, expected, *actual)
				})

				t.Run("Only hidden items", func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					_, sessionID := setup.LoggedIn(setup.User(loggedInRole))
					seller := setup.Seller()
					setup.Item(seller.UserID, aux.WithPriceInCents(100), aux.WithHidden(false))
					setup.Item(seller.UserID, aux.WithPriceInCents(200), aux.WithHidden(false))
					item3 := setup.Item(seller.UserID, aux.WithPriceInCents(400), aux.WithHidden(true))

					url := path.Items().AddHidden(true)
					request := CreateGetRequest(url, WithSessionCookie(sessionID))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code, writer.Body)

					expected := rest.ListItemsSuccessResponse{
						Items: []rest.ListItemsItemData{*FromModel(item3)},
						ListItemsStatistics: rest.ListItemsStatistics{
							TotalItemCount: 1,
							TotalItemValue: item3.PriceInCents,
						},
					}
					actual := FromJSON[rest.ListItemsSuccessResponse](t, writer.Body.String())
					require.Equal(t, expected, *actual)
				})

				t.Run("With row range", func(t *testing.T) {
					for _, limit := range []int{1, 2, 10} {
						testLabel := fmt.Sprintf("Limit %d", limit)
						t.Run(testLabel, func(t *testing.T) {
							setup, router, writer := NewRestFixture(WithDefaultCategories)
							defer setup.Close()

							itemCount := 100

							_, sessionID := setup.LoggedIn(setup.User(loggedInRole))
							seller := setup.Seller()
							items := setup.Items(seller.UserID, itemCount, aux.WithHidden(false))

							url := path.Items().AddLimit(limit)
							request := CreateGetRequest(url, WithSessionCookie(sessionID))
							router.ServeHTTP(writer, request)

							require.Equal(t, http.StatusOK, writer.Code)

							expectedItems := items[:limit]
							response := FromJSON[rest.ListItemsSuccessResponse](t, writer.Body.String())
							actualItems := response.Items
							require.Len(t, actualItems, limit)
							require.Equal(t, itemCount, response.TotalItemCount)
							require.Equal(t, aux.ItemsTotalWorth(items), response.TotalItemValue)

							for i := range limit {
								require.Equal(t, expectedItems[i].ItemID, actualItems[i].ItemID)
							}
						})
					}

					for _, offset := range []int{0, 1, 2, 10} {
						testLabel := fmt.Sprintf("Offset %d", offset)
						t.Run(testLabel, func(t *testing.T) {
							setup, router, writer := NewRestFixture(WithDefaultCategories)
							defer setup.Close()

							itemCount := 100

							_, sessionID := setup.LoggedIn(setup.User(loggedInRole))
							seller := setup.Seller()
							items := setup.Items(seller.UserID, itemCount, aux.WithHidden(false))

							url := path.Items().AddOffset(offset)
							request := CreateGetRequest(url, WithSessionCookie(sessionID))
							router.ServeHTTP(writer, request)

							require.Equal(t, http.StatusOK, writer.Code)

							expectedItems := items[offset:]
							response := FromJSON[rest.ListItemsSuccessResponse](t, writer.Body.String())
							actualItems := response.Items
							require.Len(t, actualItems, len(expectedItems))
							require.Equal(t, itemCount, response.TotalItemCount)
							require.Equal(t, aux.ItemsTotalWorth(items), response.TotalItemValue)

							for i := range len(expectedItems) - offset {
								require.Equal(t, expectedItems[i].ItemID, actualItems[i].ItemID)
							}
						})
					}

					for _, limit := range []int{1, 2, 10, 25} {
						for _, offset := range []int{0, 1, 2, 10, 25} {
							testLabel := fmt.Sprintf("Offset %d, limit %d", offset, limit)
							t.Run(testLabel, func(t *testing.T) {
								setup, router, writer := NewRestFixture(WithDefaultCategories)
								defer setup.Close()

								itemCount := 100

								_, sessionID := setup.LoggedIn(setup.User(loggedInRole))
								seller := setup.Seller()
								items := setup.Items(seller.UserID, itemCount, aux.WithHidden(false))

								url := path.Items().AddLimit(limit).AddOffset(offset)
								request := CreateGetRequest(url, WithSessionCookie(sessionID))
								router.ServeHTTP(writer, request)

								require.Equal(t, http.StatusOK, writer.Code)

								expectedItems := items[offset : offset+limit]
								response := FromJSON[rest.ListItemsSuccessResponse](t, writer.Body.String())
								actualItems := response.Items
								require.Len(t, actualItems, len(expectedItems))
								require.Equal(t, itemCount, response.TotalItemCount)
								require.Equal(t, aux.ItemsTotalWorth(items), response.TotalItemValue)

								for i := range len(expectedItems) - offset {
									require.Equal(t, expectedItems[i].ItemID, actualItems[i].ItemID)
								}
							})
						}
					}
				})

				t.Run("Filter on category", func(t *testing.T) {
					setup, router, writer := NewRestFixture()
					defer setup.Close()

					setup.Category(1, "a")
					setup.Category(2, "b")
					setup.Category(3, "c")

					_, sessionID := setup.LoggedIn(setup.User(loggedInRole))
					seller := setup.Seller()
					setup.Items(seller.UserID, 1, aux.WithHidden(false), aux.WithItemCategory(1), aux.WithPriceInCents(50))
					setup.Items(seller.UserID, 2, aux.WithHidden(false), aux.WithItemCategory(2), aux.WithPriceInCents(100))
					setup.Items(seller.UserID, 3, aux.WithHidden(false), aux.WithItemCategory(3), aux.WithPriceInCents(200))

					url := path.Items().AddCategoryFilter(2)
					request := CreateGetRequest(url, WithSessionCookie(sessionID))
					router.ServeHTTP(writer, request)

					require.Equal(t, http.StatusOK, writer.Code)
					response := FromJSON[rest.ListItemsSuccessResponse](t, writer.Body.String())

					actualItems := response.Items
					require.Len(t, actualItems, 2)
					for _, actualItem := range actualItems {
						require.Equal(t, models.ID(2), actualItem.CategoryID)
					}

					require.Equal(t, 2, response.TotalItemCount)
					require.Equal(t, models.MoneyInCents(200), response.TotalItemValue)
				})

				t.Run("Filter on description", func(t *testing.T) {
					t.Run("Searching for a", func(t *testing.T) {
						setup, router, writer := NewRestFixture(WithDefaultCategories)
						defer setup.Close()

						_, sessionID := setup.LoggedIn(setup.User(loggedInRole))
						seller := setup.Seller()
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("foo"), aux.WithPriceInCents(50))
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("bar"), aux.WithPriceInCents(100))
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("baz"), aux.WithPriceInCents(200))
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("qux"), aux.WithPriceInCents(400))

						url := path.Items().AddQueryParameter("description", "a")
						request := CreateGetRequest(url, WithSessionCookie(sessionID))
						router.ServeHTTP(writer, request)

						require.Equal(t, http.StatusOK, writer.Code)
						response := FromJSON[rest.ListItemsSuccessResponse](t, writer.Body.String())

						actualItems := response.Items
						require.Len(t, actualItems, 2)
						require.Equal(t, "bar", actualItems[0].Description)
						require.Equal(t, "baz", actualItems[1].Description)

						require.Equal(t, 2, response.ListItemsStatistics.TotalItemCount)
						require.Equal(t, models.MoneyInCents(100+200), response.ListItemsStatistics.TotalItemValue)
					})

					t.Run("Searching for space", func(t *testing.T) {
						setup, router, writer := NewRestFixture(WithDefaultCategories)
						defer setup.Close()

						_, sessionID := setup.LoggedIn(setup.User(loggedInRole))
						seller := setup.Seller()
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("foo bar"), aux.WithPriceInCents(50))
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("bar"), aux.WithPriceInCents(100))
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("baz qux"), aux.WithPriceInCents(200))
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("qux"), aux.WithPriceInCents(400))

						url := path.Items().AddQueryParameter("description", " ")
						request := CreateGetRequest(url, WithSessionCookie(sessionID))
						router.ServeHTTP(writer, request)

						require.Equal(t, http.StatusOK, writer.Code)
						response := FromJSON[rest.ListItemsSuccessResponse](t, writer.Body.String())

						actualItems := response.Items
						require.Len(t, actualItems, 2)
						require.Equal(t, "foo bar", actualItems[0].Description)
						require.Equal(t, "baz qux", actualItems[1].Description)

						require.Equal(t, 2, response.ListItemsStatistics.TotalItemCount)
						require.Equal(t, models.MoneyInCents(50+200), response.ListItemsStatistics.TotalItemValue)
					})

					t.Run("Searching for &", func(t *testing.T) {
						setup, router, writer := NewRestFixture(WithDefaultCategories)
						defer setup.Close()

						_, sessionID := setup.LoggedIn(setup.User(loggedInRole))
						seller := setup.Seller()
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("foo & bar"), aux.WithPriceInCents(50))
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("bar"), aux.WithPriceInCents(100))
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("baz & qux"), aux.WithPriceInCents(200))
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("qux&"), aux.WithPriceInCents(400))
						setup.Item(seller.UserID, aux.WithHidden(false), aux.WithDescription("qux qux"), aux.WithPriceInCents(800))

						url := path.Items().AddDescription("&")
						request := CreateGetRequest(url, WithSessionCookie(sessionID))
						router.ServeHTTP(writer, request)

						require.Equal(t, http.StatusOK, writer.Code)
						response := FromJSON[rest.ListItemsSuccessResponse](t, writer.Body.String())

						actualItems := response.Items
						require.Len(t, actualItems, 3)
						require.Equal(t, "foo & bar", actualItems[0].Description)
						require.Equal(t, "baz & qux", actualItems[1].Description)
						require.Equal(t, "qux&", actualItems[2].Description)

						require.Equal(t, 3, response.ListItemsStatistics.TotalItemCount)
						require.Equal(t, models.MoneyInCents(50+200+400), response.ListItemsStatistics.TotalItemValue)
					})
				})
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("As seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionID := setup.LoggedIn(setup.Seller())

			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
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
