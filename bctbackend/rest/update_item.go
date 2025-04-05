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

type UpdateItemData struct {
	Description  *string              `json:"description"`
	PriceInCents *models.MoneyInCents `json:"priceInCents"`
	CategoryId   *models.Id           `json:"categoryId"`
	Donation     *bool                `json:"donation"`
	Charity      *bool                `json:"charity"`
	Frozen       bool                 `json:"frozen"`
}

type UpdateItemSuccessResponse struct {
}

// @Summary Update an item.
// @Description Updates the details of an item. Only accessible to the owner of the item or an admin.
// @Tags items
// @Accept json
// @Produce json
// @Success 204 {object} UpdateItemSuccessResponse "Items successfully updated"
// @Failure 500 {object} failure_response.FailureResponse "Failed to update item"
// @Router /items/{id} [put]
func UpdateItem(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	var uriParameters struct {
		ItemId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		failure_response.BadRequest(context, err.Error())
		return
	}

	itemId, err := models.ParseId(uriParameters.ItemId)
	if err != nil {
		failure_response.BadRequest(context, err.Error())
		return
	}

	item, err := queries.GetItemWithId(db, itemId)
	if err != nil {
		var noSuchItemError *queries.NoSuchItemError
		if errors.As(err, &noSuchItemError) {
			failure_response.InvalidItemId(context, err.Error())
			return
		}
		failure_response.Unknown(context, err.Error())
		return
	}

	if !(roleId == models.AdminRoleId || (roleId == models.SellerRoleId && item.SellerId == userId)) {
		failure_response.Forbidden(context, "Only the owner of the item or an admin can update it")
		return
	}

	var payload UpdateItemData
	if err := context.ShouldBindJSON(&payload); err != nil {
		failure_response.BadRequest(context, err.Error())
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
		{
			var noSuchItemError *queries.NoSuchItemError
			if errors.As(err, &noSuchItemError) {
				failure_response.InvalidItemId(context, err.Error())
				return
			}
		}
		{
			var itemFrozenError *queries.ItemFrozenError
			if errors.As(err, &itemFrozenError) {
				failure_response.Forbidden(context, err.Error())
				return
			}
		}
		{
			var invalidPriceError *queries.InvalidPriceError
			if errors.As(err, &invalidPriceError) {
				failure_response.BadRequest(context, err.Error())
				return
			}
		}

		failure_response.Unknown(context, err.Error())
	}

	context.JSON(http.StatusNoContent, nil)
}
