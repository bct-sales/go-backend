package rest

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	rest "bctbackend/rest/shared"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetSaleInformationSuccessResponse struct {
	SaleId          models.Id          `json:"saleId" binding:"required"`
	CashierId       models.Id          `json:"cashierId" binding:"required"`
	TransactionTime rest.DateTime      `json:"transactionTime" binding:"required"`
	Items           []*GetSaleItemData `json:"items" binding:"required"`
}

type GetSaleItemData struct {
	ItemId       models.Id           `json:"itemId" binding:"required"`
	SellerId     models.Id           `json:"sellerId" binding:"required"`
	Description  string              `json:"description" binding:"required"`
	PriceInCents models.MoneyInCents `json:"priceInCents" binding:"required"`
	CategoryId   models.Id           `json:"categoryId" binding:"required"`
	Charity      *bool               `json:"charity" binding:"required"`
	Donation     *bool               `json:"donation" binding:"required"`
	AddedAt      rest.DateTime       `json:"addedAt" binding:"required"`
}

type getSaleInformationEndpoint struct {
	context *gin.Context
	db      *sql.DB
	userId  models.Id
	roleId  models.RoleId
}

func GetSaleInformation(context *gin.Context, configuration *Configuration, db *sql.DB, userId models.Id, roleId models.RoleId) {
	endpoint := &getSaleInformationEndpoint{
		context: context,
		db:      db,
		userId:  userId,
		roleId:  roleId,
	}

	endpoint.execute()
}

func (endpoint *getSaleInformationEndpoint) execute() {
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
			failure_response.UnknownSale(endpoint.context, err.Error())
			return
		}

		failure_response.Unknown(endpoint.context, "Could not retrieve sale information: "+err.Error())
		return
	}

	if endpoint.roleId.IsCashier() && sale.CashierID != endpoint.userId {
		failure_response.Forbidden(endpoint.context, "wrong_sale", "Only accessible to cashiers and owning cashiers")
		return
	}

	saleItems, err := queries.GetSaleItems(endpoint.db, saleId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchSale) {
			failure_response.UnknownSale(endpoint.context, err.Error())
			return
		}

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
		failure_response.Forbidden(endpoint.context, "wrong_role", "Only accessible to cashiers and owning cashiers")
		return false
	}

	return true
}

func (endpoint *getSaleInformationEndpoint) extractSaleIdFromUri() (models.Id, bool) {
	var uriParameters struct {
		SaleId string `uri:"id" binding:"required"`
	}
	if err := endpoint.context.ShouldBindUri(&uriParameters); err != nil {
		failure_response.InvalidUriParameters(endpoint.context, "Invalid URI parameters: "+err.Error())
		return 0, false
	}

	saleId, err := models.ParseId(uriParameters.SaleId)
	if err != nil {
		failure_response.InvalidSaleId(endpoint.context, err.Error())
		return 0, false
	}

	return saleId, true
}
