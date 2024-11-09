package seller

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type getItemsUriParameters struct {
	sellerId string `uri:"id" binding:"required"`
}

// @Summary Get seller's items
// @Description Get a seller's items
// @Produce json
// @Success 200 {object} []models.Item
// @Router /items [get]
func getItems(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.SellerRoleId {
		context.JSON(http.StatusForbidden, gin.H{"message": "Only accessible to sellers"})
		return
	}

	var uriParameters getItemsUriParameters

	if err := context.ShouldBindUri(&uriParameters); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid URI parameters"})
		return
	}

	uriSellerId, err := models.ParseId(uriParameters.sellerId)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid URI parameters"})
	}

	if userId != uriSellerId {
		context.JSON(http.StatusForbidden, gin.H{"message": "Only accessible to the seller"})
		return
	}

	items, err := queries.GetSellerItems(db, uriSellerId)

	if err != nil {
		context.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	context.IndentedJSON(http.StatusOK, items)
}
