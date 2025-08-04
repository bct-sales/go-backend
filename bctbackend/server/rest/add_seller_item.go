package rest

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/configuration"
	"bctbackend/server/failure_response"
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
func AddSellerItem(context *gin.Context, configuration *configuration.Configuration, db *sql.DB, userId models.Id, roleId models.RoleId) {
	if !roleId.IsSeller() {
		slog.Warn("Blocked attempt to add item with wrong role; front end should prevent this", "userId", userId, "roleId", roleId)
		failure_response.WrongRole(context, "Must be seller to add item")
		return
	}

	var uriParameters struct {
		SellerId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		slog.Warn("Failed to parse URI parameters", "error", err, "uriParameters", uriParameters)
		failure_response.InvalidUriParameters(context, err.Error())
		return
	}

	uriSellerId, err := models.ParseId(uriParameters.SellerId)
	if err != nil {
		slog.Warn("Failed to parse seller ID in URI", "error", err, "sellerId", uriParameters.SellerId)
		failure_response.InvalidUserId(context, err.Error())
		return
	}

	{
		sellerExists, err := queries.UserWithIdExists(db, uriSellerId)
		if err != nil {
			slog.Error("Failed to check if seller exists", "error", err, "sellerId", uriSellerId)
			failure_response.Unknown(context, err.Error())
			return
		}
		if !sellerExists {
			slog.Warn("Blocked attempt to add item for non-existing seller", "sellerId", uriSellerId)
			failure_response.UnknownUser(context, "Seller does not exist")
			return
		}
	}

	if uriSellerId != userId {
		slog.Warn("Blocked attempt to add item for different seller", "uriSellerId", uriSellerId, "userId", userId)
		failure_response.WrongSeller(context, "Logged in user does not match URI seller ID")
		return
	}

	var payload AddSellerItemPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Failed to parse AddSellerItem payload", "error", err, "payload", payload)
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
		if errors.Is(err, dberr.ErrNoSuchCategory) {
			slog.Warn("Blocked attempt to add item with unknown category", "categoryId", payload.CategoryId)
			failure_response.UnknownCategory(context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrNoSuchUser) {
			slog.Warn("Blocked attempt to add item for non-existing user", "userId", userId)
			failure_response.UnknownUser(context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrWrongRole) {
			slog.Error("[BUG] Failed to add item to non-seller; this error should have been caught earlier")
			failure_response.Unknown(context, "Bug: this error should not happen")
			return
		}

		if errors.Is(err, dberr.ErrInvalidPrice) {
			slog.Warn("Blocked attempt to add item with invalid price", "price", payload.Price)
			failure_response.InvalidPrice(context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrInvalidItemDescription) {
			slog.Warn("Blocked attempt to add item with invalid description", "description", payload.Description)
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
