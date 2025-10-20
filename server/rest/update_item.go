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

	if !ep.isOperationAuthorized(item) {
		return
	}

	if !ep.performItemUpdate(item, payload) {
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
		logger.AddInformation("error", err)
		logger.InvalidInput("Invalid URI parameters")

		failure_response.InvalidRequest(context, err.Error())

		return 0, false
	}

	itemID, err := models.ParseID(uriParameters.ItemID)
	if err != nil {
		logger.AddInformation("itemID", uriParameters.ItemID)
		logger.AddInformation("error", err)
		logger.InvalidInput("Invalid item ID in URI")

		failure_response.InvalidItemID(context, err.Error())
		return 0, false
	}

	logger.AddInformation("itemID", itemID)

	return itemID, true
}

func (ep *updateItemEndpoint) fetchItemFromDatabase(itemID models.ID) (*models.Item, bool) {
	db := ep.Database
	logger := ep.Logger
	context := ep.Context

	item, err := queries.GetItemWithID(db, itemID)
	if err != nil {
		logger.AddInformation("error", err)

		if errors.Is(err, dberr.ErrNoSuchItem) {
			logger.InvalidRequest("No such item")
			failure_response.UnknownItem(context, err.Error())
			return nil, false
		}

		logger.InternalError("Could not retrieve item")
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
		logger.AddInformation("error", err)
		logger.InvalidInput("Invalid update data")

		failure_response.InvalidRequest(context, err.Error())

		return nil, false
	}

	logger.AddInformation("payload", payload)

	return &payload, true
}

func (ep *updateItemEndpoint) isOperationAuthorized(item *models.Item) bool {
	roleID := ep.RoleID
	userID := ep.UserID
	logger := ep.Logger
	context := ep.Context

	if roleID.IsSeller() && item.SellerID != userID {
		logger.AddInformation("itemID", item.ItemID)
		logger.AddInformation("ownerID", item.SellerID)
		logger.InvalidRequest("Unauthorized item update attempt")

		failure_response.WrongSeller(context, "Only the owner of the item can update it")

		return false
	}

	if !roleID.IsAdmin() && !roleID.IsSeller() {
		logger.InvalidRequest("Unauthorized role for item update")
		failure_response.WrongRole(context, "Must be seller or admin to update item")
		return false
	}

	return true
}

func (ep *updateItemEndpoint) performItemUpdate(item *models.Item, payload *UpdateItemPayload) bool {
	db := ep.Database
	logger := ep.Logger
	context := ep.Context
	itemID := item.ItemID

	transaction, transactionErr := db.StartTransaction()
	if transactionErr != nil {
		logger.AddInformation("error", transactionErr)
		logger.InternalError("Failed to begin transaction for item update")

		failure_response.Unknown(context, transactionErr.Error())

		return false
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
		logger.AddInformation("error", updateErr)

		if errors.Is(updateErr, dberr.ErrNoSuchItem) {
			logger.AddInformation("itemID", itemID)
			logger.AddInformation("description", payload.Description)
			logger.AddInformation("priceInCents", payload.PriceInCents)
			logger.AddInformation("categoryID", payload.CategoryID)
			logger.AddInformation("donation", payload.Donation)
			logger.AddInformation("charity", payload.Charity)
			logger.InternalError("Failed to update item")

			failure_response.UnknownItem(context, updateErr.Error())

			return false
		}
		if errors.Is(updateErr, dberr.ErrItemFrozen) {
			logger.InvalidRequest("Cannot update frozen item")

			failure_response.CannotUpdateFrozenItem(context, updateErr.Error())

			return false
		}
		if errors.Is(updateErr, dberr.ErrInvalidPrice) {
			logger.InvalidInput("Invalid price in item update")
			failure_response.InvalidPrice(context, updateErr.Error())
			return false
		}

		logger.InternalError("Failed to update item")
		failure_response.Unknown(context, updateErr.Error())
	}

	if commitErr := transaction.Commit(); commitErr != nil {
		logger.AddInformation("error", commitErr)
		logger.InternalError("Failed to commit transaction after item update")
		failure_response.Unknown(context, fmt.Sprintf("Failed to commit item update: %s", commitErr.Error()))
		return false
	}

	return true
}
