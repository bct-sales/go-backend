//go:build test

package rest

import (
	"cmp"
	"net/http"
	"slices"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/database/queries"
	path "bctbackend/server/paths"
	"bctbackend/server/rest"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestCategoryCounts(t *testing.T) {
	defaultCategoryNameTable := aux.DefaultCategoryNameTable()

	t.Run("Success", func(t *testing.T) {
		t.Run("No hidden items involved", func(t *testing.T) {
			url := path.CategoriesWithCounts(queries.AllItems)

			t.Run("Zero items", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(setup.Admin())

				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				countMap := map[models.ID]int{}
				expectedResponse := createSuccessResponse(countMap)
				actualResponse := FromJson[rest.ListCategoriesSuccessResponse](t, writer.Body.String())

				require.Equal(t, len(expectedResponse.Categories), len(actualResponse.Categories))
				for i := 0; i < len(expectedResponse.Categories); i++ {
					expectedCount := expectedResponse.Categories[i]
					actualCount := actualResponse.Categories[i]

					require.Equal(t, expectedCount.CategoryID, actualCount.CategoryID)
					require.Equal(t, expectedCount.CategoryName, actualCount.CategoryName)
					require.Equal(t, expectedCount.Count, actualCount.Count)
				}
			})

			for categoryID, _ := range defaultCategoryNameTable {
				t.Run("Single item", func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					_, sessionID := setup.LoggedIn(setup.Admin())
					seller := setup.Seller()
					setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithItemCategory(categoryID), aux.WithHidden(false))

					request := CreateGetRequest(url, WithSessionCookie(sessionID))
					router.ServeHTTP(writer, request)
					countMap := map[models.ID]int{categoryID: 1}
					expected := createSuccessResponse(countMap)

					actual := FromJson[rest.ListCategoriesSuccessResponse](t, writer.Body.String())
					require.Equal(t, expected, *actual)
				})
			}

			for categoryID := range defaultCategoryNameTable {
				t.Run("Two items in same category", func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					_, sessionID := setup.LoggedIn(setup.Admin())
					seller := setup.Seller()
					setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithItemCategory(categoryID), aux.WithHidden(false))
					setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithItemCategory(categoryID), aux.WithHidden(false))

					request := CreateGetRequest(url, WithSessionCookie(sessionID))
					router.ServeHTTP(writer, request)
					countMap := map[models.ID]int{categoryID: 2}
					expected := createSuccessResponse(countMap)

					actual := FromJson[rest.ListCategoriesSuccessResponse](t, writer.Body.String())
					require.Equal(t, expected, *actual)
				})
			}

			for categoryID1 := range defaultCategoryNameTable {
				for categoryID2 := range defaultCategoryNameTable {
					if categoryID1 != categoryID2 {
						t.Run("Two items in different categories", func(t *testing.T) {
							setup, router, writer := NewRestFixture(WithDefaultCategories)
							defer setup.Close()

							_, sessionID := setup.LoggedIn(setup.Admin())
							seller := setup.Seller()
							setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithItemCategory(categoryID1), aux.WithFrozen(false), aux.WithHidden(false))
							setup.Item(seller.UserID, aux.WithDummyData(2), aux.WithItemCategory(categoryID2), aux.WithFrozen(false), aux.WithHidden(false))

							request := CreateGetRequest(url, WithSessionCookie(sessionID))
							router.ServeHTTP(writer, request)
							countMap := map[models.ID]int{categoryID1: 0, categoryID2: 0}
							countMap[categoryID1] += 1
							countMap[categoryID2] += 1
							expected := createSuccessResponse(countMap)

							actual := FromJson[rest.ListCategoriesSuccessResponse](t, writer.Body.String())
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

				_, sessionID := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				category := aux.CategoryID_BabyChildEquipment
				setup.Items(seller.UserID, 5, aux.WithItemCategory(category), aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller.UserID, 3, aux.WithItemCategory(category), aux.WithFrozen(false), aux.WithHidden(true))

				url := path.CategoriesWithCounts(queries.AllItems)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				countMap := map[models.ID]int{category: 8}
				expected := createSuccessResponse(countMap)

				actual := FromJson[rest.ListCategoriesSuccessResponse](t, writer.Body.String())
				require.NotNil(t, actual)
				require.Equal(t, expected, *actual)
			})

			t.Run("Count only hidden items", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				category := aux.CategoryID_BabyChildEquipment
				setup.Items(seller.UserID, 5, aux.WithItemCategory(category), aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller.UserID, 3, aux.WithItemCategory(category), aux.WithFrozen(false), aux.WithHidden(true))

				url := path.CategoriesWithCounts(queries.OnlyHiddenItems)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				countMap := map[models.ID]int{category: 3}
				expected := createSuccessResponse(countMap)

				actual := FromJson[rest.ListCategoriesSuccessResponse](t, writer.Body.String())
				require.NotNil(t, actual)
				require.Equal(t, expected, *actual)
			})

			t.Run("Count only visible items", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				category := aux.CategoryID_BabyChildEquipment
				setup.Items(seller.UserID, 5, aux.WithItemCategory(category), aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller.UserID, 3, aux.WithItemCategory(category), aux.WithFrozen(false), aux.WithHidden(true))

				url := path.CategoriesWithCounts(queries.OnlyVisibleItems)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				countMap := map[models.ID]int{category: 5}
				expected := createSuccessResponse(countMap)

				actual := FromJson[rest.ListCategoriesSuccessResponse](t, writer.Body.String())
				require.NotNil(t, actual)
				require.Equal(t, expected, *actual)
			})
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Not logged in", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			url := path.CategoriesWithCounts(queries.OnlyVisibleItems)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Wrong role: cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionID := setup.LoggedIn(setup.Cashier())

			url := path.CategoriesWithCounts(queries.OnlyVisibleItems)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
		})

		t.Run("Expired session", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionID := setup.LoggedIn(setup.Admin(), aux.WithExpiration(100))
			setup.Clock.Advance(500) // Advance time to ensure session is expired

			url := path.CategoriesWithCounts(queries.OnlyVisibleItems)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
		})
	})
}

func createSuccessResponse(countMap map[models.ID]int) rest.ListCategoriesSuccessResponse {
	defaultCategoryNameTable := aux.DefaultCategoryNameTable()
	countArray := []rest.CategoryData{}

	for categoryID, categoryName := range defaultCategoryNameTable {
		count, ok := countMap[categoryID]

		if !ok {
			count = 0
		}

		countArray = append(countArray, rest.CategoryData{
			CategoryID:   categoryID,
			CategoryName: categoryName,
			Count:        &count,
		})
	}

	slices.SortFunc(countArray, func(a, b rest.CategoryData) int {
		return cmp.Compare(a.CategoryID, b.CategoryID)
	})

	return rest.ListCategoriesSuccessResponse{Categories: countArray}
}
