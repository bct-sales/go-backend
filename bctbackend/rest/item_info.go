package rest

import (
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

type GetItemInformationSuccessResponse struct {
	ItemId       models.Id           `json:"itemId" binding:"required"`
	AddedAt      rest.DateTime       `json:"addedAt" binding:"required"`
	SellerId     models.Id           `json:"sellerId" binding:"required"`
	Description  string              `json:"description" binding:"required"`
	PriceInCents models.MoneyInCents `json:"priceInCents" binding:"required"`
	CategoryId   models.Id           `json:"categoryId" binding:"required"`
	Charity      *bool               `json:"charity" binding:"required"`
	Donation     *bool               `json:"donation" binding:"required"`
	Frozen       *bool               `json:"frozen" binding:"required"`
	SoldIn       *[]models.Id        `json:"soldIn" binding:"required"`
}

// @Summary Get information about an item
// @Description Get information about an item.
// @Success 200 {object} GetItemInformationSuccessResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to cashiers, admins and owner sellers"
// @Failure 404 {object} failure_response.FailureResponse "Item not found"
// @Router /items/{id} [get]
func GetItemInformation(context *gin.Context, db *sql.DB, userId models.Id, roleId models.RoleId) {
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

	item, err := queries.GetItemWithId(db, itemId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchItem) {
			failure_response.UnknownItem(context, err.Error())
			return
		}

		failure_response.Unknown(context, err.Error())
		return
	}

	if item.SellerID != userId && roleId.IsSeller() {
		failure_response.WrongSeller(context, "Only the owning seller can access this item")
		return
	}

	soldIn, err := queries.GetSalesWithItem(db, itemId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchItem) {
			failure_response.Unknown(context, "Bug: this should be caught by the previous query")
			return
		}

		failure_response.Unknown(context, err.Error())
		return
	}

	response := GetItemInformationSuccessResponse{
		ItemId:       item.ItemID,
		AddedAt:      rest.ConvertTimestampToDateTime(item.AddedAt),
		SellerId:     item.SellerID,
		Description:  item.Description,
		PriceInCents: item.PriceInCents,
		CategoryId:   item.CategoryID,
		Charity:      &item.Charity,
		Donation:     &item.Donation,
		Frozen:       &item.Frozen,
		SoldIn:       &soldIn,
	}

	context.JSON(http.StatusOK, response)
}
