package rest

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AddSellerItemPayload struct {
	Price       *models.MoneyInCents `json:"priceInCents" binding:"required"`
	Description *string              `json:"description" binding:"required"`
	CategoryId  models.Id            `json:"categoryId" binding:"required"`
	Donation    *bool                `json:"donation" binding:"required"` // needs to be a pointer to differentiate between false and not present
	Charity     *bool                `json:"charity" binding:"required"`  // needs to be a pointer to differentiate between false and not present
}

type AddSellerItemResponse struct {
	ItemId models.Id `json:"itemId"`
}

// @Summary Add an item as seller
// @Description Add an item as a seller
// @Param seller_id path int true "Seller ID"
// @Produce json
// @Success 200 {object} AddSellerItemResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to sellers and admins, or invalid item data"
// @Failure 404 {object} failure_response.FailureResponse "No such user or category"
// @Failure 500 {object} failure_response.FailureResponse "Failed to add item"
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

	{
		sellerExists, err := queries.UserWithIdExists(db, uriSellerId)
		if err != nil {
			failure_response.Unknown(context, err.Error())
			return
		}
		if !sellerExists {
			failure_response.UnknownUser(context, "Seller does not exist")
			return
		}
	}

	if uriSellerId != userId {
		failure_response.WrongSeller(context, "Logged in user does not match URI seller ID")
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
		*payload.Description,
		*payload.Price,
		payload.CategoryId,
		userId,
		*payload.Donation,
		*payload.Charity,
		false,
		false,
	)

	if err != nil {
		if errors.Is(err, database.ErrNoSuchCategory) {
			failure_response.UnknownCategory(context, err.Error())
			return
		}

		if errors.Is(err, database.ErrNoSuchUser) {
			failure_response.UnknownUser(context, err.Error())
			return
		}

		if errors.Is(err, database.ErrInvalidRole) {
			slog.Error("[BUG] Failed to add item to non-seller; this error should have been caught earlier")
			failure_response.Unknown(context, "Bug: this error should not happen")
			return
		}

		if errors.Is(err, database.ErrInvalidPrice) {
			failure_response.InvalidPrice(context, err.Error())
			return
		}

		if errors.Is(err, database.ErrInvalidItemDescription) {
			failure_response.InvalidItemDescription(context, err.Error())
			return
		}

		slog.Error("Failed to add seller item", "error", err)
		failure_response.Unknown(context, err.Error())
		return
	}

	response := AddSellerItemResponse{ItemId: itemId}
	context.JSON(http.StatusCreated, response)
}
