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
	Items []models.ID `binding:"required" json:"itemIds"`
}

type AddSaleSuccessResponse struct {
	SaleID models.ID `json:"saleId"`
}

func AddSale(arguments *HandlerFunctionArguments) {
	endpoint := addSaleEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

type addSaleEndpoint struct {
	Endpoint
}

func (ep *addSaleEndpoint) execute() {
	transaction := ep.startTransaction()
	if transaction == nil {
		return
	}
	defer transaction.RollbackIfNotCommitted()

	// Make sure user has the right role
	if !ep.ensureUserIsCashier() {
		return
	}

	// Fetch sale data
	payload := ep.parseSaleData()
	if payload == nil {
		return
	}

	saleID, saleAddedOk := ep.addSaleToDatabase(transaction, payload.Items)
	if !saleAddedOk {
		return
	}

	if !ep.endTransaction(transaction) {
		return
	}

	ep.sendSuccessResponse(saleID)
}

func (ep *addSaleEndpoint) ensureUserIsCashier() bool {
	if !ep.RoleID.IsCashier() {
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

func (ep *addSaleEndpoint) addSaleToDatabase(transaction *queries.TransactionalDatabaseQuerier, itemIDs []models.ID) (models.ID, bool) {
	timestamp := ep.Clock.Now()

	saleID, err := queries.AddSale(
		transaction,
		ep.UserID,
		timestamp,
		itemIDs,
	)
	if err != nil {
		ep.interpretDatabaseError(err)
		return 0, false
	}

	return saleID, true
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

func (ep *addSaleEndpoint) endTransaction(transaction *queries.TransactionalDatabaseQuerier) bool {
	if err := transaction.Commit(); err != nil {
		ep.Logger.InternalError("Failed to commit transaction for AddSale", "error", err)
		failure_response.Unknown(ep.Context, "Failed to commit transaction: "+err.Error())
		return false
	}

	return true
}

func (ep *addSaleEndpoint) sendSuccessResponse(saleID models.ID) {
	response := AddSaleSuccessResponse{SaleID: saleID}
	ep.Context.JSON(http.StatusCreated, response)
}
