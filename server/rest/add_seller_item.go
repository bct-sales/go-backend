package rest

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AddSellerItemPayload struct {
	Price       *models.MoneyInCents `binding:"required" json:"priceInCents"`
	Description *string              `binding:"required" json:"description"`
	CategoryID  models.ID            `binding:"required" json:"categoryId"`
	Donation    *bool                `binding:"required" json:"donation"` // needs to be a pointer to differentiate between false and not present
	Charity     *bool                `binding:"required" json:"charity"`  // needs to be a pointer to differentiate between false and not present
	Large       *bool                `binding:"required" json:"large"`    // needs to be a pointer to differentiate between false and not present
}

type AddSellerItemResponse struct {
	ItemID models.ID `json:"itemId"`
}

func AddSellerItem(context *gin.Context, arguments *HandlerFunctionArguments) {
	endpoint := addSellerItemEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(context)
}

type addSellerItemEndpoint struct {
	Endpoint
}

func (ep *addSellerItemEndpoint) execute(context *gin.Context) {
	if !ep.ensureUserIsSeller(context) {
		return
	}

	uriSellerID, parseOk := ep.parseURI(context)
	if !parseOk {
		return
	}

	if !ep.ensureValidity(context, uriSellerID) {
		return
	}

	payload := ep.parsePayload(context)
	if payload == nil {
		return
	}

	itemID, itemAddedOk := ep.addItemToDatabase(context, payload)
	if !itemAddedOk {
		return
	}

	ep.sendSuccessResponse(context, itemID)
}

func (ep *addSellerItemEndpoint) ensureUserIsSeller(context *gin.Context) bool {
	if !ep.RoleID.IsSeller() {
		ep.Logger.InvalidRequest("Blocked attempt to add item with wrong role")
		failure_response.WrongRole(context, "Must be seller to add item")
		return false
	}

	return true
}

func (ep *addSellerItemEndpoint) parseURI(context *gin.Context) (models.ID, bool) {
	logger := ep.Logger

	var uriParameters struct {
		SellerID string `binding:"required" uri:"id"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		logger.AddInformation("error", err)
		logger.AddInformation("uriParameters", uriParameters)
		logger.InvalidInput("Failed to parse URI parameters")

		failure_response.InvalidUriParameters(context, err.Error())

		return 0, false
	}

	uriSellerID, err := models.ParseID(uriParameters.SellerID)
	if err != nil {
		logger.AddInformation("sellerId", uriParameters.SellerID)
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse seller ID from URI")

		failure_response.InvalidUserID(context, err.Error())

		return 0, false
	}

	return uriSellerID, true
}

func (ep *addSellerItemEndpoint) ensureValidity(context *gin.Context, uriSellerID models.ID) bool {
	logger := ep.Logger
	ctx := context.Request.Context()

	sellerExists, err := queries.UserWithIDExists(ctx, ep.Database, uriSellerID)

	if err != nil {
		logger.AddInformation("error", err)
		logger.AddInformation("sellerId", uriSellerID)
		logger.InternalError("Failed to check if seller exists")

		failure_response.Unknown(context, err.Error())

		return false
	}

	if !sellerExists {
		logger.AddInformation("sellerId", uriSellerID)
		logger.InvalidRequest("Blocked attempt to add item for non-existing seller")

		failure_response.UnknownUser(context, "Seller does not exist")

		return false
	}

	if uriSellerID != ep.UserID {
		logger.AddInformation("sellerId", uriSellerID)
		ep.Logger.InvalidRequest("Blocked attempt to add item for different seller")

		failure_response.WrongSeller(context, "Logged in user does not match URI seller ID")

		return false
	}

	return true
}

func (ep *addSellerItemEndpoint) parsePayload(context *gin.Context) *AddSellerItemPayload {
	logger := ep.Logger

	var payload AddSellerItemPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		logger.AddInformation("error", err)
		ep.Logger.InvalidInput("Failed to parse AddSellerItem payload")

		failure_response.InvalidRequest(context, err.Error())
		return nil
	}

	ep.Logger.AddInformation("payload", payload)

	return &payload
}

func (ep *addSellerItemEndpoint) interpretDatabaseError(context *gin.Context, err error, payload *AddSellerItemPayload) {
	logger := ep.Logger

	if errors.Is(err, dberr.ErrNoSuchCategory) {
		logger.InvalidRequest("Blocked attempt to add item with unknown category")
		failure_response.UnknownCategory(context, err.Error())
		return
	}

	if errors.Is(err, dberr.ErrNoSuchUser) {
		logger.InvalidRequest("Blocked attempt to add item for non-existing user")
		failure_response.UnknownUser(context, err.Error())
		return
	}

	if errors.Is(err, dberr.ErrWrongRole) {
		logger.Bug("[BUG] Failed to add item to non-seller; this error should have been caught earlier")
		failure_response.Unknown(context, "Bug: this error should not happen")
		return
	}

	if errors.Is(err, dberr.ErrInvalidPrice) {
		logger.InvalidRequest("Blocked attempt to add item with invalid price")
		failure_response.InvalidPrice(context, err.Error())
		return
	}

	if errors.Is(err, dberr.ErrInvalidItemDescription) {
		logger.InvalidRequest("Blocked attempt to add item with invalid description")
		failure_response.InvalidItemDescription(context, err.Error())
		return
	}

	ep.Logger.InternalError("Failed to add seller item")
	failure_response.Unknown(context, err.Error())
}

func (ep *addSellerItemEndpoint) addItemToDatabase(context *gin.Context, payload *AddSellerItemPayload) (models.ID, bool) {
	ctx := context.Request.Context()
	timestamp := ep.Clock.Now()
	itemID, err := queries.AddItem(
		ctx,
		ep.Database,
		timestamp,
		*payload.Description,
		*payload.Price,
		payload.CategoryID,
		ep.UserID,
		*payload.Donation,
		*payload.Charity,
		false,
		false,
		*payload.Large,
	)

	if err != nil {
		ep.interpretDatabaseError(context, err, payload)
		return 0, false
	}

	return itemID, true
}

func (ep *addSellerItemEndpoint) sendSuccessResponse(context *gin.Context, itemID models.ID) {
	response := AddSellerItemResponse{ItemID: itemID}
	context.JSON(http.StatusCreated, response)
}
