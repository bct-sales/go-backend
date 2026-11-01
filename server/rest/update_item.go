package rest

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UpdateItemPayload struct {
	Description  *string              `json:"description"`
	PriceInCents *models.MoneyInCents `json:"priceInCents"`
	CategoryID   *models.ID           `json:"categoryId"`
	Donation     *bool                `json:"donation"`
	Charity      *bool                `json:"charity"`
	Large        *bool                `json:"large"`
}

type UpdateItemSuccessResponse struct {
}

type updateItemEndpoint struct {
	Endpoint
}

func UpdateItem(context *gin.Context, arguments *HandlerFunctionArguments) {
	endpoint := &updateItemEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(context)
}

func (ep *updateItemEndpoint) execute(context *gin.Context) {
	itemID, uriParsedOk := ep.parseUriParameters(context)
	if !uriParsedOk {
		return
	}

	// Parse payload early so that if a failure occurs, the logs contain the request data
	payload, payloadOk := ep.parsePayload(context)
	if !payloadOk {
		return
	}

	item, itemOk := ep.fetchItemFromDatabase(context, itemID)
	if !itemOk {
		return
	}

	if !ep.isOperationAuthorized(context, item) {
		return
	}

	if !ep.performItemUpdate(context, item, payload) {
		return
	}

	context.JSON(http.StatusNoContent, nil)
}

func (ep *updateItemEndpoint) parseUriParameters(context *gin.Context) (models.ID, bool) {
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

func (ep *updateItemEndpoint) fetchItemFromDatabase(context *gin.Context, itemID models.ID) (*models.Item, bool) {
	ctx := context.Request.Context()
	db := ep.Database
	logger := ep.Logger

	item, err := queries.GetItemWithID(ctx, db, itemID)
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

func (ep *updateItemEndpoint) parsePayload(context *gin.Context) (*UpdateItemPayload, bool) {
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

func (ep *updateItemEndpoint) isOperationAuthorized(context *gin.Context, item *models.Item) bool {
	roleID := ep.RoleID
	userID := ep.UserID
	logger := ep.Logger

	if !roleID.IsSeller() {
		logger.AddInformation("itemID", item.ItemID)
		logger.AddInformation("ownerID", item.SellerID)
		logger.InvalidRequest("Unauthorized item update attempt")

		failure_response.WrongRole(context, "Only the owner of the item can update it")

		return false
	}

	if item.SellerID != userID {
		logger.AddInformation("itemID", item.ItemID)
		logger.AddInformation("ownerID", item.SellerID)
		logger.InvalidRequest("Unauthorized item update attempt")

		failure_response.WrongSeller(context, "Only the owner of the item can update it")

		return false
	}

	return true
}

func (ep *updateItemEndpoint) performItemUpdate(context *gin.Context, item *models.Item, payload *UpdateItemPayload) bool {
	ctx := context.Request.Context()
	db := ep.Database
	logger := ep.Logger
	itemID := item.ItemID

	transaction, transactionErr := db.StartTransaction(ctx)
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
		Large:        payload.Large,
	}
	if updateErr := queries.UpdateItem(ctx, transaction, itemID, &itemUpdate); updateErr != nil {
		logger.AddInformation("error", updateErr)

		if errors.Is(updateErr, dberr.ErrNoSuchItem) {
			logger.AddInformation("itemID", itemID)
			logger.AddInformation("description", payload.Description)
			logger.AddInformation("priceInCents", payload.PriceInCents)
			logger.AddInformation("categoryID", payload.CategoryID)
			logger.AddInformation("donation", payload.Donation)
			logger.AddInformation("charity", payload.Charity)
			logger.AddInformation("large", payload.Large)
			logger.InternalError("Failed to update item")

			failure_response.UnknownItem(context, updateErr.Error())

			return false
		}
		if errors.Is(updateErr, dberr.ErrItemFrozen) {
			logger.InvalidRequest("Cannot update frozen item")

			failure_response.CannotUpdateFrozenItem(context, updateErr.Error())

			return false
		}
		if errors.Is(updateErr, dberr.ErrItemHidden) {
			logger.InvalidInput("Cannot update hidden item")
			failure_response.ItemHidden(context, updateErr.Error())
			return false
		}
		if errors.Is(updateErr, dberr.ErrInvalidPrice) {
			logger.InvalidInput("Invalid price in item update")
			failure_response.InvalidPrice(context, updateErr.Error())
			return false
		}

		logger.InternalError("Failed to update item")
		failure_response.Unknown(context, updateErr.Error())
		return false
	}

	if commitErr := transaction.Commit(); commitErr != nil {
		logger.AddInformation("error", commitErr)
		logger.InternalError("Failed to commit transaction after item update")
		failure_response.Unknown(context, fmt.Sprintf("Failed to commit item update: %s", commitErr.Error()))
		return false
	}

	return true
}
