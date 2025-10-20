package rest

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	"errors"
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

func AddSellerItem(arguments *HandlerFunctionArguments) {
	endpoint := addSellerItemEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

type addSellerItemEndpoint struct {
	Endpoint
}

func (ep *addSellerItemEndpoint) execute() {
	if !ep.ensureUserIsSeller() {
		return
	}

	uriSellerID, parseOk := ep.parseURI()
	if !parseOk {
		return
	}

	if !ep.ensureValidity(uriSellerID) {
		return
	}

	payload := ep.parsePayload()
	if payload == nil {
		return
	}

	itemID, itemAddedOk := ep.addItemToDatabase(payload)
	if !itemAddedOk {
		return
	}

	ep.sendSuccessResponse(itemID)
}

func (ep *addSellerItemEndpoint) ensureUserIsSeller() bool {
	if !ep.RoleID.IsSeller() {
		ep.Logger.InvalidRequest("Blocked attempt to add item with wrong role")
		failure_response.WrongRole(ep.Context, "Must be seller to add item")
		return false
	}

	return true
}

func (ep *addSellerItemEndpoint) parseURI() (models.ID, bool) {
	logger := ep.Logger

	var uriParameters struct {
		SellerID string `binding:"required" uri:"id"`
	}
	if err := ep.Context.ShouldBindUri(&uriParameters); err != nil {
		logger.AddInformation("error", err)
		logger.AddInformation("uriParameters", uriParameters)
		logger.InvalidInput("Failed to parse URI parameters")

		failure_response.InvalidUriParameters(ep.Context, err.Error())

		return 0, false
	}

	uriSellerID, err := models.ParseID(uriParameters.SellerID)
	if err != nil {
		logger.AddInformation("sellerId", uriParameters.SellerID)
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse seller ID from URI")

		failure_response.InvalidUserID(ep.Context, err.Error())

		return 0, false
	}

	return uriSellerID, true
}

func (ep *addSellerItemEndpoint) ensureValidity(uriSellerID models.ID) bool {
	logger := ep.Logger

	sellerExists, err := queries.UserWithIDExists(ep.Database, uriSellerID)

	if err != nil {
		logger.AddInformation("error", err)
		logger.AddInformation("sellerId", uriSellerID)
		logger.InternalError("Failed to check if seller exists")

		failure_response.Unknown(ep.Context, err.Error())

		return false
	}

	if !sellerExists {
		logger.AddInformation("sellerId", uriSellerID)
		logger.InvalidRequest("Blocked attempt to add item for non-existing seller")

		failure_response.UnknownUser(ep.Context, "Seller does not exist")

		return false
	}

	if uriSellerID != ep.UserID {
		logger.AddInformation("sellerId", uriSellerID)
		ep.Logger.InvalidRequest("Blocked attempt to add item for different seller")

		failure_response.WrongSeller(ep.Context, "Logged in user does not match URI seller ID")

		return false
	}

	return true
}

func (ep *addSellerItemEndpoint) parsePayload() *AddSellerItemPayload {
	logger := ep.Logger

	var payload AddSellerItemPayload
	if err := ep.Context.ShouldBindJSON(&payload); err != nil {
		logger.AddInformation("error", err)
		ep.Logger.InvalidInput("Failed to parse AddSellerItem payload")

		failure_response.InvalidRequest(ep.Context, err.Error())
		return nil
	}

	ep.Logger.AddInformation("payload", payload)

	return &payload
}

func (ep *addSellerItemEndpoint) interpretDatabaseError(err error, payload *AddSellerItemPayload) {
	logger := ep.Logger

	if errors.Is(err, dberr.ErrNoSuchCategory) {
		logger.InvalidRequest("Blocked attempt to add item with unknown category")
		failure_response.UnknownCategory(ep.Context, err.Error())
		return
	}

	if errors.Is(err, dberr.ErrNoSuchUser) {
		logger.InvalidRequest("Blocked attempt to add item for non-existing user")
		failure_response.UnknownUser(ep.Context, err.Error())
		return
	}

	if errors.Is(err, dberr.ErrWrongRole) {
		logger.Bug("[BUG] Failed to add item to non-seller; this error should have been caught earlier")
		failure_response.Unknown(ep.Context, "Bug: this error should not happen")
		return
	}

	if errors.Is(err, dberr.ErrInvalidPrice) {
		logger.InvalidRequest("Blocked attempt to add item with invalid price")
		failure_response.InvalidPrice(ep.Context, err.Error())
		return
	}

	if errors.Is(err, dberr.ErrInvalidItemDescription) {
		logger.InvalidRequest("Blocked attempt to add item with invalid description")
		failure_response.InvalidItemDescription(ep.Context, err.Error())
		return
	}

	ep.Logger.InternalError("Failed to add seller item")
	failure_response.Unknown(ep.Context, err.Error())
}

func (ep *addSellerItemEndpoint) addItemToDatabase(payload *AddSellerItemPayload) (models.ID, bool) {
	timestamp := ep.Clock.Now()
	itemID, err := queries.AddItem(
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
		ep.interpretDatabaseError(err, payload)
		return 0, false
	}

	return itemID, true
}

func (ep *addSellerItemEndpoint) sendSuccessResponse(itemID models.ID) {
	response := AddSellerItemResponse{ItemID: itemID}
	ep.Context.JSON(http.StatusCreated, response)
}
