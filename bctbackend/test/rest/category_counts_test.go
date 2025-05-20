//go:build test

package rest

import (
	"cmp"
	"net/http"
	"slices"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/rest"
	"bctbackend/rest/path"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func createSuccessResponse(countMap map[models.Id]int64) rest.ListCategoriesSuccessResponse {
	defaultCategoryTable := DefaultCategoryTable()
	countArray := []rest.CategoryData{}

	for categoryId, categoryName := range defaultCategoryTable {
		count, ok := countMap[categoryId]

		if !ok {
			count = 0
		}

		countArray = append(countArray, rest.CategoryData{
			CategoryId:   categoryId,
			CategoryName: categoryName,
			Count:        &count,
		})
	}

	slices.SortFunc(countArray, func(a, b rest.CategoryData) int {
		return cmp.Compare(a.CategoryId, b.CategoryId)
	})

	return rest.ListCategoriesSuccessResponse{Counts: countArray}
}

func TestCategoryCounts(t *testing.T) {
	url := path.Categories().WithCounts()
	defaultCategoryTable := DefaultCategoryTable()

	t.Run("Success", func(t *testing.T) {
		t.Run("Zero items", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())

			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			countMap := map[models.Id]int64{}
			expectedResponse := createSuccessResponse(countMap)
			actualResponse := FromJson[rest.ListCategoriesSuccessResponse](t, writer.Body.String())

			require.Equal(t, len(expectedResponse.Counts), len(actualResponse.Counts))
			for i := 0; i < len(expectedResponse.Counts); i++ {
				expectedCount := expectedResponse.Counts[i]
				actualCount := actualResponse.Counts[i]

				require.Equal(t, expectedCount.CategoryId, actualCount.CategoryId)
				require.Equal(t, expectedCount.CategoryName, actualCount.CategoryName)
				require.Equal(t, expectedCount.Count, actualCount.Count)
			}
		})

		for categoryId, _ := range defaultCategoryTable {
			t.Run("Single item", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionId := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				setup.Item(seller.UserId, aux.WithItemCategory(categoryId), aux.WithDummyData(1), aux.WithHidden(false))

				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				countMap := map[models.Id]int64{categoryId: 1}
				expected := createSuccessResponse(countMap)

				actual := FromJson[rest.ListCategoriesSuccessResponse](t, writer.Body.String())
				require.Equal(t, expected, *actual)
			})
		}

		for categoryId, _ := range defaultCategoryTable {
			t.Run("Two items in same category", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionId := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				setup.Item(seller.UserId, aux.WithItemCategory(categoryId), aux.WithDummyData(1), aux.WithHidden(false))
				setup.Item(seller.UserId, aux.WithItemCategory(categoryId), aux.WithDummyData(1), aux.WithHidden(false))

				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				countMap := map[models.Id]int64{categoryId: 2}
				expected := createSuccessResponse(countMap)

				actual := FromJson[rest.ListCategoriesSuccessResponse](t, writer.Body.String())
				require.Equal(t, expected, *actual)
			})
		}

		for categoryId1 := range defaultCategoryTable {
			for categoryId2 := range defaultCategoryTable {
				t.Run("Two items in potentially equal categories", func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					_, sessionId := setup.LoggedIn(setup.Admin())
					seller := setup.Seller()
					setup.Item(seller.UserId, aux.WithItemCategory(categoryId1), aux.WithDummyData(1), aux.WithHidden(false))
					setup.Item(seller.UserId, aux.WithItemCategory(categoryId2), aux.WithDummyData(2), aux.WithHidden(false))

					request := CreateGetRequest(url, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)
					countMap := map[models.Id]int64{categoryId1: 0, categoryId2: 0}
					countMap[categoryId1] += 1
					countMap[categoryId2] += 1
					expected := createSuccessResponse(countMap)

					actual := FromJson[rest.ListCategoriesSuccessResponse](t, writer.Body.String())
					require.Equal(t, expected, *actual)
				})
			}
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Not logged in", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Wrong role: cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Cashier())

			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
		})
	})
}
