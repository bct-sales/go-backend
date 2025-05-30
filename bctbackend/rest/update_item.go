package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type UpdateItemData struct {
	Description  *string              `json:"description"`
	PriceInCents *models.MoneyInCents `json:"priceInCents"`
	CategoryId   *models.Id           `json:"categoryId"`
	Donation     *bool                `json:"donation"`
	Charity      *bool                `json:"charity"`
}

type UpdateItemSuccessResponse struct {
}

// @Summary Update an item.
// @Description Updates the details of an item. Only accessible to the owner of the item or an admin.
// @Tags items
// @Accept json
// @Produce json
// @Success 204 {object} UpdateItemSuccessResponse "Items successfully updated"
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to sellers and admins, or invalid item data"
// @Failure 404 {object} failure_response.FailureResponse "Item does not exist"
// @Failure 500 {object} failure_response.FailureResponse "Failed to update item"
// @Router /items/{id} [put]
func UpdateItem(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	var uriParameters struct {
		ItemId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		failure_response.InvalidRequest(context, err.Error())
		return
	}

	itemId, err := models.ParseId(uriParameters.ItemId)
	if err != nil {
		failure_response.InvalidItemId(context, err.Error())
		return
	}

	item, err := queries.GetItemWithId(db, itemId)
	if err != nil {
		if errors.Is(err, queries.ErrNoSuchItem) {
			failure_response.UnknownItem(context, err.Error())
			return
		}
		failure_response.Unknown(context, err.Error())
		return
	}

	if roleId == models.SellerRoleId && item.SellerId != userId {
		failure_response.WrongSeller(context, "Only the owner of the item can update it")
		return
	}

	if roleId != models.AdminRoleId && roleId != models.SellerRoleId {
		failure_response.WrongRole(context, "Must be seller or admin to update item")
		return
	}

	var payload UpdateItemData
	if err := context.ShouldBindJSON(&payload); err != nil {
		failure_response.InvalidRequest(context, err.Error())
		return
	}

	itemUpdate := queries.ItemUpdate{
		Description:  payload.Description,
		PriceInCents: payload.PriceInCents,
		CategoryId:   payload.CategoryId,
		Donation:     payload.Donation,
		Charity:      payload.Charity,
	}
	if err := queries.UpdateItem(db, itemId, &itemUpdate); err != nil {
		if errors.Is(err, queries.ErrNoSuchItem) {
			slog.Error(
				"Failed to update item",
				"itemId", itemId,
				"description", payload.Description,
				"priceInCents", payload.PriceInCents,
				"categoryId", payload.CategoryId,
				"donation", payload.Donation,
				"charity", payload.Charity,
				"error", err,
			)
			failure_response.UnknownItem(context, err.Error())
			return
		}
		if errors.Is(err, queries.ItemFrozenError) {
			failure_response.CannotUpdateFrozenItem(context, err.Error())
			return
		}
		if errors.Is(err, queries.InvalidPriceError) {
			failure_response.InvalidPrice(context, err.Error())
			return
		}

		failure_response.Unknown(context, err.Error())
	}

	context.JSON(http.StatusNoContent, nil)
}
