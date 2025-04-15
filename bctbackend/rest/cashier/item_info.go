package seller

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

const (
	GetItemInformationFailureType_BadRequest    = "bad_request"
	GetItemInformationFailureType_InvalidItemId = "invalid_item_id"
	GetItemInformationFailureType_NoSuchItem    = "no_such_item"
	GetItemInformationFailureType_Unauthorized  = "unauthorized"
	GetItemInformationFailureType_Unknown       = "unknown"
)

// @Summary Get information about an item
// @Description Get information about an item.
// @Success 200 {object} GetItemInformationSuccessResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse request"
// @Failure 403 {object} failure_response.FailureResponse "Unauthorized"
// @Failure 404 {object} failure_response.FailureResponse "Item not found"
// @Router /api/v1/sales/items/{id} [get]
func GetItemInformation(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.CashierRoleId {
		failure_response.Forbidden(context, "Only accessible to cashiers")
		return
	}

	var uriParameters struct {
		ItemId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		failure_response.BadRequest(context, "Invalid URI parameters: "+err.Error())
		return
	}

	itemId, err := models.ParseId(uriParameters.ItemId)
	if err != nil {
		failure_response.BadRequest(context, "Cannot parse item Id: "+err.Error())
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
