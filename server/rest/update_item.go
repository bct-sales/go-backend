package rest

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	"errors"
	"fmt"
	"net/http"
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

func UpdateItem(arguments *HandlerFunctionArguments) {
	context := arguments.Context
	userId := arguments.UserId
	roleId := arguments.RoleId
	db := arguments.Database
	logger := arguments.Logger

	var uriParameters struct {
		ItemId string `binding:"required" uri:"id"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		logger.InvalidInput("Invalid URI parameters", "error", err)
		failure_response.InvalidRequest(context, err.Error())
		return
	}

	itemId, err := models.ParseId(uriParameters.ItemId)
	if err != nil {
		logger.InvalidInput("Invalid item ID in URI", "itemId", uriParameters.ItemId, "error", err)
		failure_response.InvalidItemId(context, err.Error())
		return
	}

	item, err := queries.GetItemWithId(db, itemId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchItem) {
			logger.InvalidRequest("No such item", "itemId", itemId, "error", err)
			failure_response.UnknownItem(context, err.Error())
			return
		}

		logger.InternalError("Could not retrieve item", "itemId", itemId, "error", err)
		failure_response.Unknown(context, err.Error())
		return
	}

	if roleId.IsSeller() && item.SellerID != userId {
		logger.InvalidRequest("Unauthorized item update attempt", "itemId", itemId, "ownerId", item.SellerID)
		failure_response.WrongSeller(context, "Only the owner of the item can update it")
		return
	}

	if !roleId.IsAdmin() && !roleId.IsSeller() {
		logger.InvalidRequest("Unauthorized role for item update", "itemId", itemId)
		failure_response.WrongRole(context, "Must be seller or admin to update item")
		return
	}

	var payload UpdateItemData
	if err := context.ShouldBindJSON(&payload); err != nil {
		logger.InvalidInput("Invalid update data", "error", err)
		failure_response.InvalidRequest(context, err.Error())
		return
	}

	transaction, err := db.StartTransaction()
	if err != nil {
		logger.InternalError("Failed to begin transaction for item update", "itemId", itemId, "error", err)
		failure_response.Unknown(context, err.Error())
		return
	}
	defer transaction.RollbackIfNotCommitted()

	itemUpdate := queries.ItemUpdate{
		AddedAt:      nil,
		Description:  payload.Description,
		PriceInCents: payload.PriceInCents,
		CategoryId:   payload.CategoryId,
		Donation:     payload.Donation,
		Charity:      payload.Charity,
	}
	if updateErr := queries.UpdateItem(transaction, itemId, &itemUpdate); updateErr != nil {
		if errors.Is(updateErr, dberr.ErrNoSuchItem) {
			logger.InternalError(
				"Failed to update item",
				"itemId", itemId,
				"description", payload.Description,
				"priceInCents", payload.PriceInCents,
				"categoryId", payload.CategoryId,
				"donation", payload.Donation,
				"charity", payload.Charity,
				"error", updateErr,
			)
			failure_response.UnknownItem(context, updateErr.Error())
			return
		}
		if errors.Is(updateErr, dberr.ErrItemFrozen) {
			logger.InvalidRequest("Cannot update frozen item", "itemId", itemId, "error", updateErr)
			failure_response.CannotUpdateFrozenItem(context, updateErr.Error())
			return
		}
		if errors.Is(updateErr, dberr.ErrInvalidPrice) {
			logger.InvalidInput("Invalid price in item update", "itemId", itemId, "priceInCents", payload.PriceInCents, "error", updateErr)
			failure_response.InvalidPrice(context, updateErr.Error())
			return
		}

		logger.InternalError("Failed to update item", "itemId", itemId, "error", updateErr)
		failure_response.Unknown(context, updateErr.Error())
	}

	if commitErr := transaction.Commit(); commitErr != nil {
		logger.InternalError("Failed to commit transaction after item update", "itemId", itemId, "error", commitErr)
		failure_response.Unknown(context, fmt.Sprintf("Failed to commit item update: %s", commitErr.Error()))
		return
	}

	context.JSON(http.StatusNoContent, nil)
}
