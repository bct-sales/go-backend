package rest

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	"errors"
	"net/http"
)

type AddSalePayload struct {
	Items []models.Id `binding:"required" json:"itemIds"`
}

type AddSaleSuccessResponse struct {
	SaleId models.Id `json:"saleId"`
}

// @Summary Add a new sale
// @Description Adds a new sale to the database. Only accessible to users with the cashier role.
// @Tags sales
// @Accept json
// @Produce json
// @Param AddSalePayload body AddSalePayload true "Payload containing item IDs"
// @Success 201 {object} AddSaleSuccessResponse "Sale successfully added"
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to cashiers"
// @Failure 404 {object} failure_response.FailureResponse "Unknown item in sale"
// @Failure 500 {object} failure_response.FailureResponse "Internal server error"
// @Router /sales [post]
func AddSale(arguments *HandlerFunctionArguments) {
	context := arguments.Context
	userId := arguments.UserId
	roleId := arguments.RoleId
	db := arguments.Database
	logger := arguments.Logger

	// Make sure user has the right role
	if !roleId.IsCashier() {
		logger.InvalidRequest("Blocked attempt to add sale with wrong role")
		failure_response.WrongRole(context, "Adding sale is only accessible to cashiers")
		return
	}

	// Fetch sale data
	var payload AddSalePayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		logger.InvalidInput("Failed to parse AddSale payload", "error", err, "payload", payload)
		failure_response.InvalidRequest(context, "Failed to parse payload:"+err.Error())
		return
	}

	// Determine current time, which will be used as the sale timestamp
	timestamp := models.Now()

	// Add the sale to the database
	saleId, err := queries.AddSale(
		db,
		userId,
		timestamp,
		payload.Items,
	)
	if err != nil {
		if errors.Is(err, dberr.ErrSaleMissingItems) {
			logger.InvalidRequest("Blocked attempt to add sale with missing items")
			failure_response.MissingItems(context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrDuplicateItemInSale) {
			logger.InvalidRequest("Blocked attempt to add sale with duplicate items")
			failure_response.DuplicateItemInSale(context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrNoSuchItem) {
			logger.InvalidRequest("Blocked attempt to add sale with unknown item; front end should prevent this")
			failure_response.UnknownItem(context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrSaleRequiresCashier) {
			logger.Bug("AddSale failed with ErrSaleRequiresCashier, but this should never occur as the role is checked before", "error", err)
			failure_response.Unknown(context, "Bug: should never occur as this is checked before")
			return
		}

		logger.InternalError("Failed to add sale", "error", err)
		failure_response.Unknown(context, "Failed to add sale: "+err.Error())
		return
	}

	response := AddSaleSuccessResponse{SaleId: saleId}
	context.JSON(http.StatusCreated, response)
}
