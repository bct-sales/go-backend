package seller

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetItemInformationResponse struct {
	Description  string              `json:"description" binding:"required"`
	PriceInCents models.MoneyInCents `json:"price_in_cents" binding:"required"`
	CategoryId   models.Id           `json:"category_id" binding:"required"`
	HasBeenSold  *bool               `json:"has_been_sold" binding:"required"`
}

func GetItemInformation(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.CashierRoleId {
		context.JSON(http.StatusForbidden, gin.H{"message": "Only accessible to cashiers"})
		return
	}

	var uriParameters struct {
		ItemId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid URI parameters: " + err.Error()})
		return
	}

	itemId, err := models.ParseId(uriParameters.ItemId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Cannot parse seller Id: " + err.Error()})
		return
	}

	saleId, err := queries.GetSaleItemInformation(db, itemId)

	// TODO Check error (e.g. not found error)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Error while querying database: " + err.Error()})
		return
	}

	hasBeenSold := saleId.SellCount > 0

	response := GetItemInformationResponse{
		Description:  saleId.Description,
		PriceInCents: saleId.PriceInCents,
		CategoryId:   saleId.ItemCategoryId,
		HasBeenSold:  &hasBeenSold,
	}

	context.JSON(http.StatusOK, response)
}
