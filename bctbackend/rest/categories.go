package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	"database/sql"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type ListCategoriesSuccessResponse struct {
	Counts []CategoryData `json:"counts"`
}

type CategoryData struct {
	CategoryId   models.Id `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	Count        *int64    `json:"count"`
}

// @Summary Get number of items grouped by category.
// @Description Returns the number of items per category.
// @Tags items
// @Accept json
// @Produce json
// @Success 200 {object} ListCategoriesSuccessResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Unauthorized access"
// @Failure 500 {object} failure_response.FailureResponse "Failed to fetch category counts"
// @Router /categories [get]
func ListCategories(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.AdminRoleId {
		failure_response.WrongRole(context, "Global category counts only accessible to admins")
		return
	}

	categoryCounts, err := queries.GetCategoryCounts(db)
	if err != nil {
		failure_response.Unknown(context, "Failed to fetch category counts: "+err.Error())
		return
	}

	response := ListCategoriesSuccessResponse{
		Counts: []CategoryData{},
	}

	for _, categoryCount := range categoryCounts {
		translatedCategoryCount := CategoryData{
			CategoryId:   categoryCount.CategoryId,
			CategoryName: categoryCount.Name,
			Count:        &categoryCount.Count,
		}

		response.Counts = append(response.Counts, translatedCategoryCount)
	}

	context.IndentedJSON(http.StatusOK, response)
}
