package admin

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type CategoryCountResponse struct {
	Counts map[string]int64 `json:"counts"`
}

// @Summary Get number of items grouped by category.
// @Description Returns the number of items per category.
// @Tags items
// @Accept json
// @Produce json
// @Success 200 {object} CategoryCountResponse"
// @Router /category-counts [get]
func GetCategoryCounts(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.AdminRoleId {
		context.JSON(http.StatusForbidden, gin.H{"message": "Only accessible to admins"})
		return
	}

	categoryMap, err := queries.GetCategoryMap(db)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch categories"})
		return
	}

	categoryCounts, err := queries.GetCategoryCounts(db)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch category counts"})
		return
	}

	response := CategoryCountResponse{
		Counts: make(map[string]int64),
	}

	for _, categoryCount := range categoryCounts {
		categoryName, ok := categoryMap[categoryCount.CategoryId]

		if !ok {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to find category name"})
			return
		}

		response.Counts[categoryName] = categoryCount.Count
	}

	context.IndentedJSON(http.StatusOK, response)
}
