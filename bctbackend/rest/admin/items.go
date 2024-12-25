package admin

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type GetItemsFailureResponse struct {
	Message string `json:"message"`
}

// @Summary Add a new sale
// @Description Returns all items. Only accessible to users with the admin role.
// @Tags items
// @Accept json
// @Produce json
// @Success 200 {object} []models.Item "Items successfully fetched"
// @Failure 500 {object} GetItemsFailureResponse "Failed to fetch items"
// @Router /items [get]
func GetItems(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.AdminRoleId {
		context.JSON(http.StatusForbidden, gin.H{"message": "Only accessible to admins"})
		return
	}

	items, err := queries.GetItems(db)

	if err != nil {
		failureResponse := GetItemsFailureResponse{Message: "Failed to fetch items"}
		context.JSON(http.StatusInternalServerError, failureResponse)
		return
	}

	context.IndentedJSON(http.StatusOK, items)
}
