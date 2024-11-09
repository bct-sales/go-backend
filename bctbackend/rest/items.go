package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

// @Summary Get all items
// @Description Get all items
// @Accept json
// @Produce json
// @Success 200 {object} []models.Item
// @Router /items [get]
func getItems(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.AdminRoleId {
		context.JSON(http.StatusForbidden, gin.H{"message": "Only accessible to admins"})
		return
	}

	items, err := queries.GetItems(db)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch items"})
		return
	}

	context.IndentedJSON(http.StatusOK, items)
}
