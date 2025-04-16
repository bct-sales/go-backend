package seller

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AddSellerItemPayload struct {
	Price       *models.MoneyInCents `json:"price_in_cents" binding:"required"`
	Description string               `json:"description" binding:"required"`
	CategoryId  models.Id            `json:"category_id" binding:"required"`
	Donation    *bool                `json:"donation" binding:"required"` // needs to be a pointer to differentiate between false and not present
	Charity     *bool                `json:"charity" binding:"required"`  // needs to be a pointer to differentiate between false and not present
}

type AddSellerItemResponse struct {
	ItemId models.Id `json:"item_id"`
}

// @Summary Add an item as seller
// @Description Add an item as a seller
// @Param seller_id path int true "Seller ID"
// @Produce json
// @Success 200 {object} AddSellerItemResponse
// @Router /seller/{seller_id}/items [put]
func AddSellerItem(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.SellerRoleId {
		failure_response.WrongRole(context, "Must be seller to add item")
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

	if uriSellerId != userId {
		failure_response.WrongUser(context, "Logged in user does not match URI seller ID")
		return
	}

	var payload AddSellerItemPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		failure_response.InvalidRequest(context, err.Error())
		return
	}

	timestamp := models.Now()

	itemId, err := queries.AddItem(
		db,
		timestamp,
		payload.Description,
		*payload.Price,
		payload.CategoryId,
		userId,
		*payload.Donation,
		*payload.Charity,
		false,
	)

	if err != nil {
		{
			var noSuchCategoryError *queries.NoSuchCategoryError
			if errors.As(err, &noSuchCategoryError) {
				failure_response.UnknownCategory(context, err.Error())
				return
			}
		}

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
				failure_response.Unknown(context, "Bug: this error should not happen")
				return
			}
		}

		{
			var invalidPriceError *queries.InvalidPriceError
			if errors.As(err, &invalidPriceError) {
				failure_response.InvalidPrice(context, err.Error())
				return
			}
		}

		failure_response.Unknown(context, err.Error())
		return
	}

	response := AddSellerItemResponse{ItemId: itemId}
	context.JSON(http.StatusCreated, response)
}
