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
	Counts map[models.Id]int64 `json:"counts"`
}

func GetCategoryCounts(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.AdminRoleId {
		context.JSON(http.StatusForbidden, gin.H{"message": "Only accessible to admins"})
		return
	}

	categoryCounts, err := queries.GetCategoryCounts(db)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch category counts"})
		return
	}

	response := CategoryCountResponse{
		Counts: make(map[models.Id]int64),
	}

	for _, categoryCount := range categoryCounts {
		response.Counts[categoryCount.CategoryId] = categoryCount.Count
	}

	context.IndentedJSON(http.StatusOK, response)
}
