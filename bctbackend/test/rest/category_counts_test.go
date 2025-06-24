//go:build test

package rest

import (
	"cmp"
	"net/http"
	"slices"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server"
	"bctbackend/server/path"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func createSuccessResponse(countMap map[models.Id]int) server.ListCategoriesSuccessResponse {
	defaultCategoryNameTable := aux.DefaultCategoryNameTable()
	countArray := []server.CategoryData{}

	for categoryId, categoryName := range defaultCategoryNameTable {
		count, ok := countMap[categoryId]

		if !ok {
			count = 0
		}

		countArray = append(countArray, server.CategoryData{
			CategoryId:   categoryId,
			CategoryName: categoryName,
			Count:        &count,
		})
	}

	slices.SortFunc(countArray, func(a, b server.CategoryData) int {
		return cmp.Compare(a.CategoryId, b.CategoryId)
	})

	return server.ListCategoriesSuccessResponse{Categories: countArray}
}

func TestCategoryCounts(t *testing.T) {
	defaultCategoryNameTable := aux.DefaultCategoryNameTable()

	t.Run("Success", func(t *testing.T) {
		t.Run("No hidden items involved", func(t *testing.T) {
			url := path.Categories().WithCounts(queries.AllItems)

			t.Run("Zero items", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionId := setup.LoggedIn(setup.Admin())

				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				countMap := map[models.Id]int{}
				expectedResponse := createSuccessResponse(countMap)
				actualResponse := FromJson[server.ListCategoriesSuccessResponse](t, writer.Body.String())

				require.Equal(t, len(expectedResponse.Categories), len(actualResponse.Categories))
				for i := 0; i < len(expectedResponse.Categories); i++ {
					expectedCount := expectedResponse.Categories[i]
					actualCount := actualResponse.Categories[i]

					require.Equal(t, expectedCount.CategoryId, actualCount.CategoryId)
					require.Equal(t, expectedCount.CategoryName, actualCount.CategoryName)
					require.Equal(t, expectedCount.Count, actualCount.Count)
				}
			})

			for categoryId, _ := range defaultCategoryNameTable {
				t.Run("Single item", func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					_, sessionId := setup.LoggedIn(setup.Admin())
					seller := setup.Seller()
					setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithItemCategory(categoryId), aux.WithHidden(false))

					request := CreateGetRequest(url, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)
					countMap := map[models.Id]int{categoryId: 1}
					expected := createSuccessResponse(countMap)

					actual := FromJson[server.ListCategoriesSuccessResponse](t, writer.Body.String())
					require.Equal(t, expected, *actual)
				})
			}

			for categoryId, _ := range defaultCategoryNameTable {
				t.Run("Two items in same category", func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					_, sessionId := setup.LoggedIn(setup.Admin())
					seller := setup.Seller()
					setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithItemCategory(categoryId), aux.WithHidden(false))
					setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithItemCategory(categoryId), aux.WithHidden(false))

					request := CreateGetRequest(url, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)
					countMap := map[models.Id]int{categoryId: 2}
					expected := createSuccessResponse(countMap)

					actual := FromJson[server.ListCategoriesSuccessResponse](t, writer.Body.String())
					require.Equal(t, expected, *actual)
				})
			}

			for categoryId1 := range defaultCategoryNameTable {
				for categoryId2 := range defaultCategoryNameTable {
					if categoryId1 != categoryId2 {
						t.Run("Two items in different categories", func(t *testing.T) {
							setup, router, writer := NewRestFixture(WithDefaultCategories)
							defer setup.Close()

							_, sessionId := setup.LoggedIn(setup.Admin())
							seller := setup.Seller()
							setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithItemCategory(categoryId1), aux.WithFrozen(false), aux.WithHidden(false))
							setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithItemCategory(categoryId2), aux.WithFrozen(false), aux.WithHidden(false))

							request := CreateGetRequest(url, WithSessionCookie(sessionId))
							router.ServeHTTP(writer, request)
							countMap := map[models.Id]int{categoryId1: 0, categoryId2: 0}
							countMap[categoryId1] += 1
							countMap[categoryId2] += 1
							expected := createSuccessResponse(countMap)

							actual := FromJson[server.ListCategoriesSuccessResponse](t, writer.Body.String())
							require.NotNil(t, actual)
							require.Equal(t, expected, *actual)
						})
					}
				}
			}
		})

		t.Run("Hidden items involved", func(t *testing.T) {
			t.Run("Count all items", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionId := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				category := aux.CategoryId_BabyChildEquipment
				setup.Items(seller.UserId, 5, aux.WithItemCategory(category), aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller.UserId, 3, aux.WithItemCategory(category), aux.WithFrozen(false), aux.WithHidden(true))

				url := path.Categories().WithCounts(queries.AllItems)
				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				countMap := map[models.Id]int{category: 8}
				expected := createSuccessResponse(countMap)

				actual := FromJson[server.ListCategoriesSuccessResponse](t, writer.Body.String())
				require.NotNil(t, actual)
				require.Equal(t, expected, *actual)
			})

			t.Run("Count only hidden items", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionId := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				category := aux.CategoryId_BabyChildEquipment
				setup.Items(seller.UserId, 5, aux.WithItemCategory(category), aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller.UserId, 3, aux.WithItemCategory(category), aux.WithFrozen(false), aux.WithHidden(true))

				url := path.Categories().WithCounts(queries.OnlyHiddenItems)
				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				countMap := map[models.Id]int{category: 3}
				expected := createSuccessResponse(countMap)

				actual := FromJson[server.ListCategoriesSuccessResponse](t, writer.Body.String())
				require.NotNil(t, actual)
				require.Equal(t, expected, *actual)
			})

			t.Run("Count only visible items", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionId := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				category := aux.CategoryId_BabyChildEquipment
				setup.Items(seller.UserId, 5, aux.WithItemCategory(category), aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller.UserId, 3, aux.WithItemCategory(category), aux.WithFrozen(false), aux.WithHidden(true))

				url := path.Categories().WithCounts(queries.OnlyVisibleItems)
				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				countMap := map[models.Id]int{category: 5}
				expected := createSuccessResponse(countMap)

				actual := FromJson[server.ListCategoriesSuccessResponse](t, writer.Body.String())
				require.NotNil(t, actual)
				require.Equal(t, expected, *actual)
			})
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Not logged in", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			url := path.Categories().WithCounts(queries.OnlyVisibleItems)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Wrong role: cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Cashier())

			url := path.Categories().WithCounts(queries.OnlyVisibleItems)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
		})
	})
}
