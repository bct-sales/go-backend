package seller

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

// @Summary Get seller's items
// @Description Get a seller's items
// @Param seller_id path int true "Seller ID"
// @Produce json
// @Success 200 {object} []models.Item
// @Router /seller/{seller_id}/items [get]
func GetSellerItems(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.SellerRoleId {
		context.JSON(http.StatusForbidden, gin.H{"message": "Only accessible to sellers"})
		return
	}

	var uriParameters struct {
		SellerId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid URI parameters: " + err.Error()})
		return
	}

	uriSellerId, err := models.ParseId(uriParameters.SellerId)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Cannot parse seller Id: " + err.Error()})
		return
	}

	if userId != uriSellerId {
		context.JSON(http.StatusForbidden, gin.H{"message": "Logged in user does not match URI seller ID"})
		return
	}

	items, err := queries.GetSellerItems(db, uriSellerId)

	if err != nil {
		context.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	context.IndentedJSON(http.StatusOK, items)
}
