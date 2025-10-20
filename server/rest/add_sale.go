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
	// Make sure user has the right role
	if !ep.ensureUserIsCashier() {
		return
	}

	// Extract sale data from request
	payload := ep.parseSaleData()
	if payload == nil {
		return
	}

	transaction := ep.startTransaction()
	if transaction == nil {
		return
	}
	defer transaction.RollbackIfNotCommitted()

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
	logger := ep.Logger

	if !ep.RoleID.IsCashier() {
		logger.InvalidRequest("Blocked attempt to add sale with wrong role")
		failure_response.WrongRole(ep.Context, "Adding sale is only accessible to cashiers")
		return false
	}

	return true
}

func (ep *addSaleEndpoint) parseSaleData() *AddSalePayload {
	logger := ep.Logger
	var payload AddSalePayload

	if err := ep.Context.ShouldBindJSON(&payload); err != nil {
		logger.AddInformation("payload", payload)
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse AddSale payload")

		failure_response.InvalidRequest(ep.Context, "Failed to parse payload:"+err.Error())

		return nil
	}

	ep.Logger.AddInformation("payload", payload)

	return &payload
}

func (ep *addSaleEndpoint) interpretDatabaseError(err error) {
	logger := ep.Logger
	logger.AddInformation("error", err)

	if errors.Is(err, dberr.ErrSaleMissingItems) {
		logger.InvalidRequest("Blocked attempt to add sale with missing items")
		failure_response.MissingItems(ep.Context, err.Error())
		return
	}

	if errors.Is(err, dberr.ErrDuplicateItemInSale) {
		logger.InvalidRequest("Blocked attempt to add sale with duplicate items")
		failure_response.DuplicateItemInSale(ep.Context, err.Error())
		return
	}

	if errors.Is(err, dberr.ErrNoSuchItem) {
		logger.InvalidRequest("Blocked attempt to add sale with unknown item; front end should prevent this")
		failure_response.UnknownItem(ep.Context, err.Error())
		return
	}

	if errors.Is(err, dberr.ErrSaleRequiresCashier) {
		logger.Bug("AddSale failed with ErrSaleRequiresCashier, but this should never occur as the role is checked before")
		failure_response.Unknown(ep.Context, "Bug: should never occur as this is checked before")
		return
	}

	logger.InternalError("Failed to add sale")
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
	logger := ep.Logger
	transaction, err := ep.Database.StartTransaction()

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to start transaction for AddSale")

		failure_response.Unknown(ep.Context, "Failed to start transaction: "+err.Error())

		return nil
	}

	return transaction
}

func (ep *addSaleEndpoint) endTransaction(transaction *queries.TransactionalDatabaseQuerier) bool {
	logger := ep.Logger

	if err := transaction.Commit(); err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to commit transaction for AddSale")

		failure_response.Unknown(ep.Context, "Failed to commit transaction: "+err.Error())

		return false
	}

	return true
}

func (ep *addSaleEndpoint) sendSuccessResponse(saleID models.ID) {
	response := AddSaleSuccessResponse{SaleID: saleID}
	ep.Context.JSON(http.StatusCreated, response)
}
