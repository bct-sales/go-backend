//go:build test

package rest

import (
	"net/http"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/defs"
	"bctbackend/rest"
	"bctbackend/rest/path"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func createSuccessResponse(countMap map[models.Id]int64) rest.ListCategoriesSuccessResponse {
	countArray := []rest.CategoryData{}

	for _, categoryId := range defs.ListCategories() {
		count, ok := countMap[categoryId]

		if !ok {
			count = 0
		}

		categoryName, err := defs.NameOfCategory(categoryId)

		if err != nil {
			panic(err)
		}

		countArray = append(countArray, rest.CategoryData{
			CategoryId:   categoryId,
			CategoryName: categoryName,
			Count:        count,
		})
	}

	return rest.ListCategoriesSuccessResponse{Counts: countArray}
}

func TestCategoryCounts(t *testing.T) {
	url := path.CategoryCounts().String()

	t.Run("Success", func(t *testing.T) {
		t.Run("Zero items", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())

			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			countMap := map[models.Id]int64{}
			expectedResponse := createSuccessResponse(countMap)
			actual := FromJson[rest.ListCategoriesSuccessResponse](writer.Body.String())
			require.Equal(t, expectedResponse, *actual)
		})

		for _, categoryId := range defs.ListCategories() {
			t.Run("Single item", func(t *testing.T) {
				setup, router, writer := NewRestFixture()
				defer setup.Close()

				_, sessionId := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				setup.Item(seller.UserId, aux.WithItemCategory(categoryId), aux.WithDummyData(1))

				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				countMap := map[models.Id]int64{categoryId: 1}
				expected := createSuccessResponse(countMap)

				actual := FromJson[rest.ListCategoriesSuccessResponse](writer.Body.String())
				require.Equal(t, expected, *actual)
			})
		}

		for _, categoryId := range defs.ListCategories() {
			t.Run("Two items in same category", func(t *testing.T) {
				setup, router, writer := NewRestFixture()
				defer setup.Close()

				_, sessionId := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				setup.Item(seller.UserId, aux.WithItemCategory(categoryId), aux.WithDummyData(1))
				setup.Item(seller.UserId, aux.WithItemCategory(categoryId), aux.WithDummyData(1))

				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				countMap := map[models.Id]int64{categoryId: 2}
				expected := createSuccessResponse(countMap)

				actual := FromJson[rest.ListCategoriesSuccessResponse](writer.Body.String())
				require.Equal(t, expected, *actual)
			})
		}

		for _, categoryId1 := range defs.ListCategories() {
			for _, categoryId2 := range defs.ListCategories() {
				t.Run("Two items in potentially equal categories", func(t *testing.T) {
					setup, router, writer := NewRestFixture()
					defer setup.Close()

					_, sessionId := setup.LoggedIn(setup.Admin())
					seller := setup.Seller()
					setup.Item(seller.UserId, aux.WithItemCategory(categoryId1), aux.WithDummyData(1))
					setup.Item(seller.UserId, aux.WithItemCategory(categoryId2), aux.WithDummyData(2))

					request := CreateGetRequest(url, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)
					countMap := map[models.Id]int64{categoryId1: 0, categoryId2: 0}
					countMap[categoryId1] += 1
					countMap[categoryId2] += 1
					expected := createSuccessResponse(countMap)

					actual := FromJson[rest.ListCategoriesSuccessResponse](writer.Body.String())
					require.Equal(t, expected, *actual)
				})
			}
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Not logged in", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Wrong role", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Seller())

			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
		})
	})
}
