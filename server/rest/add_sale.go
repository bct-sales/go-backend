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

type AddSalePayload struct {
	Items []models.ID `binding:"required" json:"itemIds"`
}

type AddSaleSuccessResponse struct {
	SaleID models.ID `json:"saleId"`
}

func AddSale(context *gin.Context, arguments *HandlerFunctionArguments) {
	endpoint := addSaleEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(context)
}

type addSaleEndpoint struct {
	Endpoint
}

func (ep *addSaleEndpoint) execute(context *gin.Context) {
	// Make sure user has the right role
	if !ep.ensureUserIsCashier(context) {
		return
	}

	// Extract sale data from request
	payload := ep.parseSaleData(context)
	if payload == nil {
		return
	}

	transaction := ep.startTransaction(context)
	if transaction == nil {
		return
	}
	defer transaction.RollbackIfNotCommitted()

	saleID, saleAddedOk := ep.addSaleToDatabase(context, transaction, payload.Items)
	if !saleAddedOk {
		return
	}

	if !ep.endTransaction(context, transaction) {
		return
	}

	ep.sendSuccessResponse(context, saleID)
}

func (ep *addSaleEndpoint) ensureUserIsCashier(context *gin.Context) bool {
	logger := ep.Logger

	if !ep.RoleID.IsCashier() {
		logger.InvalidRequest("Blocked attempt to add sale with wrong role")
		failure_response.WrongRole(context, "Adding sale is only accessible to cashiers")
		return false
	}

	return true
}

func (ep *addSaleEndpoint) parseSaleData(context *gin.Context) *AddSalePayload {
	logger := ep.Logger
	var payload AddSalePayload

	if err := context.ShouldBindJSON(&payload); err != nil {
		logger.AddInformation("payload", payload)
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse AddSale payload")

		failure_response.InvalidRequest(context, "Failed to parse payload:"+err.Error())

		return nil
	}

	ep.Logger.AddInformation("payload", payload)

	return &payload
}

func (ep *addSaleEndpoint) interpretDatabaseError(context *gin.Context, err error) {
	logger := ep.Logger
	logger.AddInformation("error", err)

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
		logger.Bug("AddSale failed with ErrSaleRequiresCashier, but this should never occur as the role is checked before")
		failure_response.Unknown(context, "Bug: should never occur as this is checked before")
		return
	}

	logger.InternalError("Failed to add sale")
	failure_response.Unknown(context, "Failed to add sale: "+err.Error())
}

func (ep *addSaleEndpoint) addSaleToDatabase(context *gin.Context, transaction *queries.TransactionalDatabaseQuerier, itemIDs []models.ID) (models.ID, bool) {
	ctx := context.Request.Context()
	timestamp := ep.Clock.Now()

	saleID, err := queries.AddSale(
		ctx,
		transaction,
		ep.UserID,
		timestamp,
		itemIDs,
	)
	if err != nil {
		ep.interpretDatabaseError(context, err)
		return 0, false
	}

	return saleID, true
}

func (ep *addSaleEndpoint) startTransaction(context *gin.Context) *queries.TransactionalDatabaseQuerier {
	logger := ep.Logger
	ctx := context.Request.Context()
	transaction, err := ep.Database.StartTransaction(ctx)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to start transaction for AddSale")

		failure_response.Unknown(context, "Failed to start transaction: "+err.Error())

		return nil
	}

	return transaction
}

func (ep *addSaleEndpoint) endTransaction(context *gin.Context, transaction *queries.TransactionalDatabaseQuerier) bool {
	logger := ep.Logger

	if err := transaction.Commit(); err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to commit transaction for AddSale")

		failure_response.Unknown(context, "Failed to commit transaction: "+err.Error())

		return false
	}

	return true
}

func (ep *addSaleEndpoint) sendSuccessResponse(context *gin.Context, saleID models.ID) {
	response := AddSaleSuccessResponse{SaleID: saleID}
	context.JSON(http.StatusCreated, response)
}
