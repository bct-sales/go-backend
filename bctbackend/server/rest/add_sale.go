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
	endpoint := addSaleEndpoint{
		HandlerFunctionArguments: *arguments,
	}

	endpoint.execute()
}

type addSaleEndpoint struct {
	HandlerFunctionArguments
}

func (ep *addSaleEndpoint) execute() {
	transaction := ep.startTransaction()
	if transaction == nil {
		return
	}
	defer transaction.Rollback()

	// Make sure user has the right role
	if !ep.ensureUserIsCashier() {
		return
	}

	// Fetch sale data
	payload := ep.parseSaleData()
	if payload == nil {
		return
	}

	saleId, saleIdOk := ep.addSaleToDatabase(transaction, payload.Items)
	if !saleIdOk {
		return
	}

	if !ep.EndTransaction(transaction) {
		return
	}

	ep.sendResponse(saleId)
}

func (ep *addSaleEndpoint) ensureUserIsCashier() bool {
	if !ep.RoleId.IsCashier() {
		ep.Logger.InvalidRequest("Blocked attempt to add sale with wrong role")
		failure_response.WrongRole(ep.Context, "Adding sale is only accessible to cashiers")
		return false
	}

	return true
}

func (ep *addSaleEndpoint) parseSaleData() *AddSalePayload {
	var payload AddSalePayload

	if err := ep.Context.ShouldBindJSON(&payload); err != nil {
		ep.Logger.InvalidInput("Failed to parse AddSale payload", "error", err, "payload", payload)
		failure_response.InvalidRequest(ep.Context, "Failed to parse payload:"+err.Error())
		return nil
	}

	return &payload
}

func (ep *addSaleEndpoint) interpretDatabaseError(err error) {
	if errors.Is(err, dberr.ErrSaleMissingItems) {
		ep.Logger.InvalidRequest("Blocked attempt to add sale with missing items")
		failure_response.MissingItems(ep.Context, err.Error())
		return
	}

	if errors.Is(err, dberr.ErrDuplicateItemInSale) {
		ep.Logger.InvalidRequest("Blocked attempt to add sale with duplicate items")
		failure_response.DuplicateItemInSale(ep.Context, err.Error())
		return
	}

	if errors.Is(err, dberr.ErrNoSuchItem) {
		ep.Logger.InvalidRequest("Blocked attempt to add sale with unknown item; front end should prevent this")
		failure_response.UnknownItem(ep.Context, err.Error())
		return
	}

	if errors.Is(err, dberr.ErrSaleRequiresCashier) {
		ep.Logger.Bug("AddSale failed with ErrSaleRequiresCashier, but this should never occur as the role is checked before", "error", err)
		failure_response.Unknown(ep.Context, "Bug: should never occur as this is checked before")
		return
	}

	ep.Logger.InternalError("Failed to add sale", "error", err)
	failure_response.Unknown(ep.Context, "Failed to add sale: "+err.Error())
}

func (ep *addSaleEndpoint) addSaleToDatabase(transaction *queries.TransactionalDatabaseQuerier, itemIDs []models.Id) (models.Id, bool) {
	timestamp := ep.Clock.Now()

	saleId, err := queries.AddSale(
		transaction,
		ep.UserId,
		timestamp,
		itemIDs,
	)
	if err != nil {
		ep.interpretDatabaseError(err)
		return 0, false
	}

	return saleId, true
}

func (ep *addSaleEndpoint) startTransaction() *queries.TransactionalDatabaseQuerier {
	transaction, err := ep.Database.StartTransaction()

	if err != nil {
		ep.Logger.InternalError("Failed to start transaction for AddSale", "error", err)
		failure_response.Unknown(ep.Context, "Failed to start transaction: "+err.Error())
		return nil
	}

	return transaction
}

func (ep *addSaleEndpoint) EndTransaction(transaction *queries.TransactionalDatabaseQuerier) bool {
	if err := transaction.Commit(); err != nil {
		ep.Logger.InternalError("Failed to commit transaction for AddSale", "error", err)
		failure_response.Unknown(ep.Context, "Failed to commit transaction: "+err.Error())
		return false
	}

	return true
}

func (ep *addSaleEndpoint) sendResponse(saleId models.Id) {
	response := AddSaleSuccessResponse{SaleId: saleId}
	ep.Context.JSON(http.StatusCreated, response)
}
