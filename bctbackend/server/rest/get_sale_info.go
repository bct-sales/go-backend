package rest

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	"bctbackend/server/logger"
	rest "bctbackend/server/shared"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetSaleInformationSuccessResponse struct {
	SaleId          models.Id          `binding:"required" json:"saleId"`
	CashierId       models.Id          `binding:"required" json:"cashierId"`
	TransactionTime rest.DateTime      `binding:"required" json:"transactionTime"`
	Items           []*GetSaleItemData `binding:"required" json:"items"`
}

type GetSaleItemData struct {
	ItemId       models.Id           `binding:"required" json:"itemId"`
	SellerId     models.Id           `binding:"required" json:"sellerId"`
	Description  string              `binding:"required" json:"description"`
	PriceInCents models.MoneyInCents `binding:"required" json:"priceInCents"`
	CategoryId   models.Id           `binding:"required" json:"categoryId"`
	Charity      *bool               `binding:"required" json:"charity"`
	Donation     *bool               `binding:"required" json:"donation"`
	AddedAt      rest.DateTime       `binding:"required" json:"addedAt"`
}

type getSaleInformationEndpoint struct {
	context *gin.Context
	db      *sql.DB
	userId  models.Id
	roleId  models.RoleId
	logger  logger.Logger
}

func GetSaleInformation(arguments *HandlerFunctionArguments) {
	endpoint := &getSaleInformationEndpoint{
		context: arguments.Context,
		db:      arguments.Database,
		userId:  arguments.UserId,
		roleId:  arguments.RoleId,
		logger:  arguments.Logger,
	}

	endpoint.execute()
}

func (endpoint *getSaleInformationEndpoint) execute() {
	logger := endpoint.logger

	if !endpoint.ensureUserHasRightRole() {
		return
	}

	saleId, ok := endpoint.extractSaleIdFromUri()
	if !ok {
		return
	}

	sale, err := queries.GetSaleWithId(endpoint.db, saleId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchSale) {
			logger.InvalidRequest("No such sale found", "saleId", saleId)
			failure_response.UnknownSale(endpoint.context, err.Error())
			return
		}

		logger.InternalError("Could not retrieve sale information", "saleId", saleId, "error", err)
		failure_response.Unknown(endpoint.context, "Could not retrieve sale information: "+err.Error())
		return
	}

	if endpoint.roleId.IsCashier() && sale.CashierID != endpoint.userId {
		logger.InvalidRequest("Sale is not owned by the cashier", "saleId", saleId, "saleOwnerId", sale.CashierID)
		failure_response.Forbidden(endpoint.context, "wrong_sale", "Only accessible to cashiers and owning cashiers")
		return
	}

	saleItems, err := queries.GetSaleItems(endpoint.db, saleId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchSale) {
			logger.Bug("No such sale found; should have been caught earlier", "saleId", saleId)
			failure_response.UnknownSale(endpoint.context, err.Error())
			return
		}

		logger.InternalError("Could not retrieve sale items", "saleId", saleId, "error", err)
		failure_response.Unknown(endpoint.context, "Could not retrieve sale information: "+err.Error())
		return
	}

	response := GetSaleInformationSuccessResponse{
		SaleId:          sale.SaleID,
		CashierId:       sale.CashierID,
		TransactionTime: rest.ConvertTimestampToDateTime(sale.TransactionTime),
		Items:           algorithms.Map(saleItems, endpoint.convertSaleItemToData),
	}

	endpoint.context.JSON(http.StatusOK, response)
}

func (endpoint *getSaleInformationEndpoint) convertSaleItemToData(saleItem *models.Item) *GetSaleItemData {
	return &GetSaleItemData{
		ItemId:       saleItem.ItemID,
		SellerId:     saleItem.SellerID,
		Description:  saleItem.Description,
		PriceInCents: saleItem.PriceInCents,
		CategoryId:   saleItem.CategoryID,
		Charity:      &saleItem.Charity,
		Donation:     &saleItem.Donation,
		AddedAt:      rest.ConvertTimestampToDateTime(saleItem.AddedAt),
	}
}

func (endpoint *getSaleInformationEndpoint) ensureUserHasRightRole() bool {
	if !endpoint.roleId.IsAdmin() && !endpoint.roleId.IsCashier() {
		endpoint.logger.InvalidRequest("User does not have the right role to access sale information")
		failure_response.Forbidden(endpoint.context, "wrong_role", "Only accessible to cashiers and owning cashiers")
		return false
	}

	return true
}

func (endpoint *getSaleInformationEndpoint) extractSaleIdFromUri() (models.Id, bool) {
	var uriParameters struct {
		SaleId string `binding:"required" uri:"id"`
	}
	if err := endpoint.context.ShouldBindUri(&uriParameters); err != nil {
		endpoint.logger.InvalidInput("Invalid URI parameters", "error", err)
		failure_response.InvalidUriParameters(endpoint.context, "Invalid URI parameters: "+err.Error())
		return 0, false
	}

	saleId, err := models.ParseId(uriParameters.SaleId)
	if err != nil {
		endpoint.logger.InvalidInput("Invalid sale ID", "saleId", uriParameters.SaleId, "error", err)
		failure_response.InvalidSaleId(endpoint.context, err.Error())
		return 0, false
	}

	return saleId, true
}
