package seller

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
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
	if roleId != models.SellerRoleId && roleId != models.AdminRoleId {
		failure_response.Forbidden(context, "wrong_role", "Only accessible to sellers and admins")
		return
	}

	var uriParameters struct {
		SellerId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		failure_response.InvalidUriParameters(context, err.Error())
		return
	}

	uriSellerId, err := models.ParseId(uriParameters.SellerId)
	if err != nil {
		failure_response.InvalidUserId(context, err.Error())
		return
	}

	if userId != uriSellerId && roleId != models.AdminRoleId {
		failure_response.WrongSeller(context, "Logged in user does not match URI seller ID")
		return
	}

	items, err := queries.GetSellerItems(db, uriSellerId)
	if err != nil {
		failure_response.Unknown(context, "Could not retrieve seller items: "+err.Error())
		return
	}

	context.IndentedJSON(http.StatusOK, items)
}
