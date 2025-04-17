package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	"database/sql"
	"errors"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

// @Summary Get seller's items
// @Description Get a seller's items
// @Param seller_id path int true "Seller ID"
// @Produce json
// @Success 200 {object} []models.Item
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to owning sellers and admins"
// @Failure 404 {object} failure_response.FailureResponse "No such user"
// @Failure 500 {object} failure_response.FailureResponse "Failed to fetch items"
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

	if err := queries.CheckUserRole(db, uriSellerId, models.SellerRoleId); err != nil {
		{
			var noSuchUserError *queries.NoSuchUserError
			if errors.As(err, &noSuchUserError) {
				failure_response.UnknownUser(context, err.Error())
				return
			}
		}

		{
			var invalidRoleError *queries.InvalidRoleError
			if errors.As(err, &invalidRoleError) {
				failure_response.WrongUser(context, "Can only list items of sellers")
				return
			}
		}

		failure_response.Unknown(context, "Could not check user role: "+err.Error())
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
