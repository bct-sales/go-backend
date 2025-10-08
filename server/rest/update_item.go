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

type UpdateItemPayload struct {
	Description  *string              `json:"description"`
	PriceInCents *models.MoneyInCents `json:"priceInCents"`
	CategoryID   *models.ID           `json:"categoryId"`
	Donation     *bool                `json:"donation"`
	Charity      *bool                `json:"charity"`
}

type UpdateItemSuccessResponse struct {
}

type updateItemEndpoint struct {
	Endpoint
}

func UpdateItem(arguments *HandlerFunctionArguments) {
	endpoint := &updateItemEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

func (ep *updateItemEndpoint) execute() {
	context := ep.Context
	db := ep.Database
	logger := ep.Logger

	itemID, uriParsedOk := ep.parseUriParameters()
	if !uriParsedOk {
		return
	}

	// Parse payload early so that if a failure occurs, the logs contain the request data
	payload, payloadOk := ep.parsePayload()
	if !payloadOk {
		return
	}

	item, itemOk := ep.fetchItemFromDatabase(itemID)
	if !itemOk {
		return
	}

	if !ep.isUpdateAuthorized(item) {
		return
	}

	transaction, err := db.StartTransaction()
	if err != nil {
		logger.InternalError("Failed to begin transaction for item update", "itemID", itemID, "error", err)
		failure_response.Unknown(context, err.Error())
		return
	}
	defer transaction.RollbackIfNotCommitted()

	itemUpdate := queries.ItemUpdate{
		AddedAt:      nil,
		Description:  payload.Description,
		PriceInCents: payload.PriceInCents,
		CategoryID:   payload.CategoryID,
		Donation:     payload.Donation,
		Charity:      payload.Charity,
	}
	if updateErr := queries.UpdateItem(transaction, itemID, &itemUpdate); updateErr != nil {
		if errors.Is(updateErr, dberr.ErrNoSuchItem) {
			logger.InternalError(
				"Failed to update item",
				"itemID", itemID,
				"description", payload.Description,
				"priceInCents", payload.PriceInCents,
				"categoryID", payload.CategoryID,
				"donation", payload.Donation,
				"charity", payload.Charity,
				"error", updateErr,
			)
			failure_response.UnknownItem(context, updateErr.Error())
			return
		}
		if errors.Is(updateErr, dberr.ErrItemFrozen) {
			logger.InvalidRequest("Cannot update frozen item", "itemID", itemID, "error", updateErr)
			failure_response.CannotUpdateFrozenItem(context, updateErr.Error())
			return
		}
		if errors.Is(updateErr, dberr.ErrInvalidPrice) {
			logger.InvalidInput("Invalid price in item update", "itemID", itemID, "priceInCents", payload.PriceInCents, "error", updateErr)
			failure_response.InvalidPrice(context, updateErr.Error())
			return
		}

		logger.InternalError("Failed to update item", "itemID", itemID, "error", updateErr)
		failure_response.Unknown(context, updateErr.Error())
	}

	if commitErr := transaction.Commit(); commitErr != nil {
		logger.InternalError("Failed to commit transaction after item update", "itemID", itemID, "error", commitErr)
		failure_response.Unknown(context, fmt.Sprintf("Failed to commit item update: %s", commitErr.Error()))
		return
	}

	context.JSON(http.StatusNoContent, nil)
}

func (ep *updateItemEndpoint) parseUriParameters() (models.ID, bool) {
	context := ep.Context
	logger := ep.Logger

	var uriParameters struct {
		ItemID string `binding:"required" uri:"id"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		logger.InvalidInput("Invalid URI parameters", "error", err)
		failure_response.InvalidRequest(context, err.Error())
		return 0, false
	}

	itemID, err := models.ParseID(uriParameters.ItemID)
	if err != nil {
		logger.InvalidInput("Invalid item ID in URI", "itemID", uriParameters.ItemID, "error", err)
		failure_response.InvalidItemID(context, err.Error())
		return 0, false
	}

	return itemID, true
}

func (ep *updateItemEndpoint) fetchItemFromDatabase(itemID models.ID) (*models.Item, bool) {
	db := ep.Database
	logger := ep.Logger
	context := ep.Context

	item, err := queries.GetItemWithID(db, itemID)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchItem) {
			logger.InvalidRequest("No such item", "itemID", itemID, "error", err)
			failure_response.UnknownItem(context, err.Error())
			return nil, false
		}

		logger.InternalError("Could not retrieve item", "itemID", itemID, "error", err)
		failure_response.Unknown(context, err.Error())
		return nil, false
	}

	return item, true
}

func (ep *updateItemEndpoint) parsePayload() (*UpdateItemPayload, bool) {
	context := ep.Context
	logger := ep.Logger

	var payload UpdateItemPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		logger.InvalidInput("Invalid update data", "error", err)
		failure_response.InvalidRequest(context, err.Error())
		return nil, false
	}

	logger.AddInformation("payload", payload)

	return &payload, true
}

func (ep *updateItemEndpoint) isUpdateAuthorized(item *models.Item) bool {
	roleID := ep.RoleID
	userID := ep.UserID
	logger := ep.Logger
	context := ep.Context

	if roleID.IsSeller() && item.SellerID != userID {
		logger.InvalidRequest("Unauthorized item update attempt", "itemID", item.ItemID, "ownerID", item.SellerID)
		failure_response.WrongSeller(context, "Only the owner of the item can update it")
		return false
	}

	if !roleID.IsAdmin() && !roleID.IsSeller() {
		logger.InvalidRequest("Unauthorized role for item update", "itemID", item.ItemID)
		failure_response.WrongRole(context, "Must be seller or admin to update item")
		return false
	}

	return true
}
