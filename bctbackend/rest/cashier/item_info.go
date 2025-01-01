package seller

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
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

type GetItemInformationFailureResponse struct {
	Type    string `json:"type"`
	Details string `json:"details"`
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
// @Failure 400 {object} GetItemInformationFailureResponse "Failed to parse request"
// @Failure 403 {object} GetItemInformationFailureResponse "Unauthorized"
// @Failure 404 {object} GetItemInformationFailureResponse "Item not found"
// @Router /api/v1/sales/items/{id} [get]
func GetItemInformation(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.CashierRoleId {
		failureResponse := GetItemInformationFailureResponse{Type: GetItemInformationFailureType_Unauthorized, Details: "Only accessible to cashiers"}
		context.JSON(http.StatusForbidden, failureResponse)
		return
	}

	var uriParameters struct {
		ItemId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		failureResponse := GetItemInformationFailureResponse{Type: GetItemInformationFailureType_BadRequest, Details: err.Error()}
		context.JSON(http.StatusBadRequest, failureResponse)
		return
	}

	itemId, err := models.ParseId(uriParameters.ItemId)
	if err != nil {
		failureResponse := GetItemInformationFailureResponse{Type: GetItemInformationFailureType_InvalidItemId, Details: err.Error()}
		context.JSON(http.StatusBadRequest, failureResponse)
		return
	}

	saleId, err := queries.GetSaleItemInformation(db, itemId)

	if err != nil {
		var itemNotFoundError *queries.ItemNotFoundError
		if errors.As(err, &itemNotFoundError) {
			failureResponse := GetItemInformationFailureResponse{Type: GetItemInformationFailureType_NoSuchItem, Details: err.Error()}
			context.JSON(http.StatusNotFound, failureResponse)
		}

		failureResponse := GetItemInformationFailureResponse{Type: GetItemInformationFailureType_Unknown, Details: err.Error()}
		context.JSON(http.StatusInternalServerError, failureResponse)
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
