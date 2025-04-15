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

type CategoryCountSuccessResponse struct {
	Counts []CategoryCount `json:"counts"`
}

type CategoryCount struct {
	CategoryId   models.Id `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Count        int64     `json:"count"`
}

// @Summary Get number of items grouped by category.
// @Description Returns the number of items per category.
// @Tags items
// @Accept json
// @Produce json
// @Success 200 {object} CategoryCountSuccessResponse
// @Failure 403 {object} failure_response.FailureResponse "Unauthorized access"
// @Failure 500 {object} failure_response.FailureResponse "Failed to fetch category counts"
// @Router /category-counts [get]
func GetCategoryCounts(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.AdminRoleId {
		failure_response.WrongRole(context, "Global category counts only accessible to admins")
		return
	}

	categoryCounts, err := queries.GetCategoryCounts(db)
	if err != nil {
		failure_response.Unknown(context, "Failed to fetch category counts: "+err.Error())
		return
	}

	response := CategoryCountSuccessResponse{
		Counts: []CategoryCount{},
	}

	for _, categoryCount := range categoryCounts {
		translatedCategoryCount := CategoryCount{
			CategoryId:   categoryCount.CategoryId,
			CategoryName: categoryCount.Name,
			Count:        categoryCount.Count,
		}

		response.Counts = append(response.Counts, translatedCategoryCount)
	}

	context.IndentedJSON(http.StatusOK, response)
}
