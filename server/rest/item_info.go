package rest

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"errors"
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
	SoldIn       *[]models.ID        `binding:"required" json:"soldIn"`
}

type getItemInformationEndpoint struct {
	Endpoint
}

func GetItemInformation(arguments *HandlerFunctionArguments) {
	endpoint := getItemInformationEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

func (ep *getItemInformationEndpoint) execute() {
	itemID, foundItemID := ep.retrieveItemIDFromUri()
	if !foundItemID {
		return
	}

	item := ep.retrieveItemFromDatabase(itemID)
	if item == nil {
		return
	}

	if !ep.ensureQueryAllowed(item) {
		return
	}

	salesIncludingItem, findSalesOk := ep.findSalesIncludingItem(itemID)
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
		SoldIn:       &salesIncludingItem,
	}

	ep.Context.JSON(http.StatusOK, response)
}

func (ep *getItemInformationEndpoint) retrieveItemIDFromUri() (models.ID, bool) {
	logger := ep.Logger

	var uriParameters struct {
		ItemID string `binding:"required" uri:"id"`
	}

	if err := ep.Context.ShouldBindUri(&uriParameters); err != nil {
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse URI parameters")

		failure_response.InvalidUriParameters(ep.Context, "Invalid URI parameters: "+err.Error())

		return 0, false
	}

	itemID, err := models.ParseID(uriParameters.ItemID)
	if err != nil {
		logger.AddInformation("itemID", uriParameters.ItemID)
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse item ID")

		failure_response.InvalidItemID(ep.Context, err.Error())

		return 0, false
	}

	logger.AddInformation("itemID", itemID)

	return itemID, true
}

func (ep *getItemInformationEndpoint) retrieveItemFromDatabase(itemID models.ID) *models.Item {
	logger := ep.Logger

	item, err := queries.GetItemWithID(ep.Database, itemID)
	if err != nil {
		logger.AddInformation("error", err)

		if errors.Is(err, dberr.ErrNoSuchItem) {
			logger.InvalidRequest("Attempt to access a non-existing item")

			failure_response.UnknownItem(ep.Context, err.Error())

			return nil
		}

		logger.InternalError("Failed to fetch item from database")

		failure_response.Unknown(ep.Context, err.Error())

		return nil
	}

	logger.AddInformation("item", item)

	return item
}

func (ep *getItemInformationEndpoint) ensureQueryAllowed(item *models.Item) bool {
	logger := ep.Logger

	if ep.RoleID.IsSeller() && item.SellerID != ep.UserID {
		logger.InvalidRequest("Blocked attempt to access item not owned by the seller")

		failure_response.WrongSeller(ep.Context, "Only the owning seller can access this item")

		return false
	}

	if ep.RoleID.IsCashier() && item.Hidden {
		logger.InvalidRequest("Blocked attempt to access hidden item by cashier")

		failure_response.ItemHidden(ep.Context, "Cashiers cannot see hidden items")

		return false
	}

	return true
}

func (ep *getItemInformationEndpoint) findSalesIncludingItem(itemID models.ID) ([]models.ID, bool) {
	logger := ep.Logger

	soldIn, err := queries.GetSalesWithItem(ep.Database, itemID)
	if err != nil {
		logger.AddInformation("error", err)

		if errors.Is(err, dberr.ErrNoSuchItem) {
			logger.Bug("Unknown item; should have been caught earlier")

			failure_response.Unknown(ep.Context, "Bug: this should be caught by the previous query")

			return nil, false
		}

		logger.InternalError("Failed to get sales for item")

		failure_response.Unknown(ep.Context, err.Error())

		return nil, false
	}

	return soldIn, true
}
