package rest

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GetItemInformationSuccessResponse struct {
	ItemID       models.ID           `binding:"required" json:"itemId"`
	AddedAt      rest.DateTime       `binding:"required" json:"addedAt"`
	SellerID     models.ID           `binding:"required" json:"sellerId"`
	Description  string              `binding:"required" json:"description"`
	PriceInCents models.MoneyInCents `binding:"required" json:"priceInCents"`
	CategoryID   models.ID           `binding:"required" json:"categoryId"`
	Charity      *bool               `binding:"required" json:"charity"`
	Donation     *bool               `binding:"required" json:"donation"`
	Frozen       *bool               `binding:"required" json:"frozen"`
	Large        *bool               `binding:"required" json:"large"`
	SoldIn       *[]models.ID        `binding:"required" json:"soldIn"`
}

type getItemInformationEndpoint struct {
	Endpoint
}

func GetItemInformation(context *gin.Context, arguments *HandlerFunctionArguments) {
	endpoint := getItemInformationEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(context)
}

func (ep *getItemInformationEndpoint) execute(context *gin.Context) {
	itemID, foundItemID := ep.retrieveItemIDFromUri(context)
	if !foundItemID {
		return
	}

	item := ep.retrieveItemFromDatabase(context, itemID)
	if item == nil {
		return
	}

	if !ep.ensureQueryAllowed(context, item) {
		return
	}

	salesIncludingItem, findSalesOk := ep.findSalesIncludingItem(context, itemID)
	if !findSalesOk {
		return
	}

	response := GetItemInformationSuccessResponse{
		ItemID:       item.ItemID,
		AddedAt:      rest.ConvertTimestampToDateTime(item.AddedAt),
		SellerID:     item.SellerID,
		Description:  item.Description,
		PriceInCents: item.PriceInCents,
		CategoryID:   item.CategoryID,
		Charity:      &item.Charity,
		Donation:     &item.Donation,
		Frozen:       &item.Frozen,
		Large:        &item.Large,
		SoldIn:       &salesIncludingItem,
	}

	context.JSON(http.StatusOK, response)
}

func (ep *getItemInformationEndpoint) retrieveItemIDFromUri(context *gin.Context) (models.ID, bool) {
	logger := ep.Logger

	var uriParameters struct {
		ItemID string `binding:"required" uri:"id"`
	}

	if err := context.ShouldBindUri(&uriParameters); err != nil {
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse URI parameters")

		failure_response.InvalidUriParameters(context, "Invalid URI parameters: "+err.Error())

		return 0, false
	}

	itemID, err := models.ParseID(uriParameters.ItemID)
	if err != nil {
		logger.AddInformation("itemID", uriParameters.ItemID)
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse item ID")

		failure_response.InvalidItemID(context, err.Error())

		return 0, false
	}

	logger.AddInformation("itemID", itemID)

	return itemID, true
}

func (ep *getItemInformationEndpoint) retrieveItemFromDatabase(context *gin.Context, itemID models.ID) *models.Item {
	logger := ep.Logger
	ctx := context.Request.Context()

	item, err := queries.GetItemWithID(ctx, ep.Database, itemID)
	if err != nil {
		logger.AddInformation("error", err)

		if errors.Is(err, dberr.ErrNoSuchItem) {
			logger.InvalidRequest("Attempt to access a non-existing item")

			failure_response.UnknownItem(context, err.Error())

			return nil
		}

		logger.InternalError("Failed to fetch item from database")

		failure_response.Unknown(context, err.Error())

		return nil
	}

	logger.AddInformation("item", item)

	return item
}

func (ep *getItemInformationEndpoint) ensureQueryAllowed(context *gin.Context, item *models.Item) bool {
	logger := ep.Logger

	if ep.RoleID.IsSeller() && item.SellerID != ep.UserID {
		logger.InvalidRequest("Blocked attempt to access item not owned by the seller")

		failure_response.WrongSeller(context, "Only the owning seller can access this item")

		return false
	}

	if ep.RoleID.IsCashier() && item.Hidden {
		logger.InvalidRequest("Blocked attempt to access hidden item by cashier")

		failure_response.ItemHidden(context, "Cashiers cannot see hidden items")

		return false
	}

	return true
}

func (ep *getItemInformationEndpoint) findSalesIncludingItem(context *gin.Context, itemID models.ID) ([]models.ID, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()

	soldIn, err := queries.GetSalesWithItem(ctx, ep.Database, itemID)
	if err != nil {
		logger.AddInformation("error", err)

		if errors.Is(err, dberr.ErrNoSuchItem) {
			logger.Bug("Unknown item; should have been caught earlier")

			failure_response.Unknown(context, "Bug: this should be caught by the previous query")

			return nil, false
		}

		logger.InternalError("Failed to get sales for item")

		failure_response.Unknown(context, err.Error())

		return nil, false
	}

	return soldIn, true
}
