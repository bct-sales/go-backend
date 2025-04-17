package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetItemInformationSuccessResponse struct {
	Description  string              `json:"description" binding:"required"`
	PriceInCents models.MoneyInCents `json:"price_in_cents" binding:"required"`
	CategoryId   models.Id           `json:"category_id" binding:"required"`
	HasBeenSold  *bool               `json:"has_been_sold" binding:"required"`
}

// @Summary Get information about an item
// @Description Get information about an item.
// @Success 200 {object} GetItemInformationSuccessResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to cashiers"
// @Failure 404 {object} failure_response.FailureResponse "Item not found"
// @Router items/{id} [get]
func GetItemInformation(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.CashierRoleId {
		failure_response.WrongRole(context, "Only accessible to cashiers")
		return
	}

	var uriParameters struct {
		ItemId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		failure_response.InvalidUriParameters(context, "Invalid URI parameters: "+err.Error())
		return
	}

	itemId, err := models.ParseId(uriParameters.ItemId)
	if err != nil {
		failure_response.InvalidItemId(context, err.Error())
		return
	}

	saleId, err := queries.GetSaleItemInformation(db, itemId)
	if err != nil {
		var NoSuchItemError *queries.NoSuchItemError
		if errors.As(err, &NoSuchItemError) {
			failure_response.UnknownItem(context, "No such item: "+err.Error())
			return
		}

		failure_response.Unknown(context, "Failed to get item information: "+err.Error())
		return
	}

	hasBeenSold := saleId.SellCount > 0

	response := GetItemInformationSuccessResponse{
		Description:  saleId.Description,
		PriceInCents: saleId.PriceInCents,
		CategoryId:   saleId.ItemCategoryId,
		HasBeenSold:  &hasBeenSold,
	}

	context.JSON(http.StatusOK, response)
}
