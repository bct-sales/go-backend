package rest

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"errors"
	"fmt"
	"net/http"
)

// GetUserInformationByAdminSuccessResponse is the common response part for all user information requests done by admins.
// It is embedded in other response structs specialized for sellers, cashiers and admins.
type GetUserInformationByAdminSuccessResponse struct {
	UserId       models.ID      `binding:"required"            json:"userId"`
	Role         string         `binding:"required"            json:"role"`
	Password     string         `binding:"required"            json:"password"`
	CreatedAt    rest.DateTime  `binding:"required"            json:"createdAt"`
	LastActivity *rest.DateTime `json:"lastActivity,omitempty"`
}

// GetSellerInformationByAdminSuccessResponse is the response struct for seller information requests done by admins.
type GetSellerInformationByAdminSuccessResponse struct {
	GetUserInformationByAdminSuccessResponse
	Items *[]*GetSellerInformationItemData `binding:"required" json:"items"`
}

// GetCashierInformationByAdminSuccessResponse is the response struct for cashier information requests done by admins.
type GetCashierInformationByAdminSuccessResponse struct {
	GetUserInformationByAdminSuccessResponse
	Sales *[]*GetCashierInformationByAdminSaleData `binding:"required" json:"sales"`
}

// GetAdminInformationByAdminSuccessResponse is the response struct for admin information requests done by admins.
type GetAdminInformationByAdminSuccessResponse struct {
	GetUserInformationByAdminSuccessResponse
}

// GetSellerInformationItemData contains data regarding items.
// It is used when an admin requests information about a seller.
type GetSellerInformationItemData struct {
	ItemId       models.ID           `binding:"required" json:"itemId"`
	AddedAt      rest.DateTime       `binding:"required" json:"addedAt"`
	SellerId     models.ID           `binding:"required" json:"sellerId"`
	Description  string              `binding:"required" json:"description"`
	PriceInCents models.MoneyInCents `binding:"required" json:"priceInCents"`
	CategoryId   models.ID           `binding:"required" json:"categoryId"`
	Charity      *bool               `binding:"required" json:"charity"`
	Donation     *bool               `binding:"required" json:"donation"`
	Frozen       *bool               `binding:"required" json:"frozen"`
	SaleCount    *int                `binding:"required" json:"saleCount"`
}

// GetCashierInformationByAdminSaleData contains data regarding sales.
// It is used when an admin requests information about a cashier.
type GetCashierInformationByAdminSaleData struct {
	SaleId          models.ID     `binding:"required" json:"saleId"`
	TransactionTime rest.DateTime `binding:"required" json:"transactionTime"`
}

// GetSellerInformationBySellerSuccessResponse is the response struct for seller information requests done by sellers.
type GetSellerInformationBySellerSuccessResponse struct {
	ItemCount       int                 `binding:"required" json:"itemCount"`
	FrozenItemCount int                 `binding:"required" json:"frozenItemCount"`
	HiddenItemCount int                 `binding:"required" json:"hiddenItemCount"`
	TotalPrice      models.MoneyInCents `binding:"required" json:"totalPrice"`
}

// GetCashierInformationByCashierSuccessResponse is the response struct for cashier information requests done by cashiers.
type GetCashierInformationByCashierSuccessResponse struct {
	Sales *[]*GetCashierInformationByAdminSaleData `binding:"required" json:"sales"`
}

func convertItemToGetUserInformationItem(item *queries.ItemWithSaleCount) *GetSellerInformationItemData {
	return &GetSellerInformationItemData{
		ItemId:       item.ItemID,
		Description:  item.Description,
		SellerId:     item.SellerID,
		AddedAt:      rest.ConvertTimestampToDateTime(item.AddedAt),
		PriceInCents: item.PriceInCents,
		CategoryId:   item.CategoryID,
		Charity:      &item.Charity,
		Donation:     &item.Donation,
		Frozen:       &item.Frozen,
		SaleCount:    &item.SaleCount,
	}
}

func convertSaleToGetUserInformationSale(sale *models.Sale) *GetCashierInformationByAdminSaleData {
	return &GetCashierInformationByAdminSaleData{
		SaleId:          sale.SaleID,
		TransactionTime: rest.ConvertTimestampToDateTime(sale.TransactionTime),
	}
}

// @Summary Get information about a user
// @Description Get information about a user.
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to admins"
// @Failure 404 {object} failure_response.FailureResponse "User not found"
// @Failure 500 {object} failure_response.FailureResponse "Internal server error"
// @Router /users/{id} [get]
func GetUserInformation(arguments *HandlerFunctionArguments) {
	endpoint := GetUserInformationEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

type GetUserInformationEndpoint struct {
	Endpoint
}

func (ep *GetUserInformationEndpoint) execute() {
	context := ep.Context
	roleId := ep.RoleID

	queriedUserId, err := ep.retrieveQueriedUserFromUri()
	if err != nil {
		return
	}

	if roleId.IsAdmin() {
		// If the user is an admin, they can access any user's information
		ep.getUserInformationAsAdmin(queriedUserId)
		return
	} else if roleId.IsSeller() {
		ep.getUserInformationAsSeller(queriedUserId)
		return
	} else if roleId.IsCashier() {
		ep.getUserInformationAsCashier(queriedUserId)
		return
	} else {
		failure_response.Unknown(context, fmt.Sprintf("Bug: unhandled role %d", roleId))
		return
	}
}

func (ep *GetUserInformationEndpoint) retrieveQueriedUserFromUri() (models.ID, error) {
	// Retrieve id of user whose information is being requested
	var uriParameters struct {
		UserId string `binding:"required" uri:"id"`
	}
	if err := ep.Context.ShouldBindUri(&uriParameters); err != nil {
		ep.Logger.InvalidInput("Invalid URI parameters", "error", err)
		failure_response.InvalidUriParameters(ep.Context, "Invalid URI parameters: "+err.Error())
		return 0, err
	}

	// Parse user id
	queriedUserId, err := models.ParseID(uriParameters.UserId)
	if err != nil {
		ep.Logger.InvalidInput("Invalid user ID", "error", err, "userId", uriParameters.UserId)
		failure_response.InvalidUserID(ep.Context, err.Error())
		return 0, err
	}

	return queriedUserId, nil
}

func (ep *GetUserInformationEndpoint) getUserInformationAsAdmin(queriedUserId models.ID) {
	logger := ep.Logger
	context := ep.Context
	db := ep.Database

	// Look up user in database
	user, err := queries.GetUserWithID(db, queriedUserId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchUser) {
			logger.InvalidRequest("User not found", "queriedUserId", queriedUserId)
			failure_response.UnknownUser(context, err.Error())
			return
		}

		logger.InternalError("Could not find user in database", "queriedUserId", queriedUserId, "error", err)
		failure_response.Unknown(context, err.Error())
		return
	}

	basicInformation := GetUserInformationByAdminSuccessResponse{
		UserId:       user.UserID,
		Role:         user.RoleID.Name(),
		Password:     user.Password,
		CreatedAt:    rest.ConvertTimestampToDateTime(user.CreatedAt),
		LastActivity: algorithms.MapOptional(user.LastActivity, rest.ConvertTimestampToDateTime),
	}

	if user.RoleID.IsAdmin() {
		response := GetAdminInformationByAdminSuccessResponse{
			GetUserInformationByAdminSuccessResponse: basicInformation,
		}
		context.JSON(http.StatusOK, response)
		return
	} else if user.RoleID.IsSeller() {
		items, err := queries.GetSellerItemsWithSaleCounts(db, user.UserID)
		if err != nil {
			{
				if errors.Is(err, dberr.ErrNoSuchUser) {
					logger.Bug("User not found; should have been caught earlier", "queriedUserId", queriedUserId)
					failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
					return
				}
			}
			if errors.Is(err, dberr.ErrWrongRole) {
				logger.Bug("User has the wrong role; should have been caught earlier", "queriedUserId", queriedUserId)
				failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
				return
			}

			logger.InternalError("Failed to get seller items with sale counts", "queriedUserId", queriedUserId, "error", err)
			failure_response.Unknown(context, fmt.Errorf("failed to find information about seller: %w", err).Error())
			return
		}

		convertedItems := algorithms.Map(items, convertItemToGetUserInformationItem)

		response := GetSellerInformationByAdminSuccessResponse{
			GetUserInformationByAdminSuccessResponse: basicInformation,
			Items:                                    &convertedItems,
		}

		context.JSON(http.StatusOK, response)
		return
	} else if user.RoleID.IsCashier() {
		sales, err := queries.GetSalesWithCashier(db, user.UserID)
		if err != nil {
			if errors.Is(err, dberr.ErrNoSuchUser) {
				logger.Bug("User not found; should have been caught earlier", "queriedUserId", queriedUserId)
				failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
				return
			}
			if errors.Is(err, dberr.ErrWrongRole) {
				logger.Bug("User has the wrong role; should have been caught earlier", "queriedUserId", queriedUserId)
				failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
				return
			}

			logger.InternalError("Failed to get sales with cashier", "queriedUserId", queriedUserId, "error", err)
			failure_response.Unknown(context, err.Error())
			return
		}

		convertedSales := algorithms.Map(sales, convertSaleToGetUserInformationSale)

		response := GetCashierInformationByAdminSuccessResponse{
			GetUserInformationByAdminSuccessResponse: basicInformation,
			Sales:                                    &convertedSales,
		}
		context.JSON(http.StatusOK, response)
		return
	} else {
		logger.Bug("Unhandled user role")
		failure_response.Unknown(context, fmt.Sprintf("Bug: unhandled role %d", user.RoleID.Int64()))
		return
	}
}

func (ep *GetUserInformationEndpoint) getUserInformationAsSeller(queriedUserId models.ID) {
	logger := ep.Logger
	context := ep.Context
	db := ep.Database
	userId := ep.UserID

	if userId != queriedUserId {
		logger.InvalidRequest("Seller attempted to access another user's information", "queriedUserId", queriedUserId)
		failure_response.WrongRole(context, "Only admins can access other users' information")
		return
	}

	itemCount, err := queries.CountSellerItems(db, queriedUserId, queries.IncludeAll, queries.Exclude)
	if err != nil {
		// At this point, we know that the user exists and is a seller, so no errors should ever occur
		logger.InternalError("Failed to count seller items", "queriedUserId", queriedUserId, "error", err)
		failure_response.Unknown(context, err.Error())
		return
	}

	frozenItemCount, err := queries.CountSellerItems(db, queriedUserId, queries.IncludeOnly, queries.IncludeAll)
	if err != nil {
		// At this point, we know that the user exists and is a seller, so no errors should ever occur
		logger.InternalError("Failed to count frozen seller items", "queriedUserId", queriedUserId, "error", err)
		failure_response.Unknown(context, err.Error())
		return
	}

	hiddenItemCount, err := queries.CountSellerItems(db, queriedUserId, queries.IncludeAll, queries.IncludeOnly)
	if err != nil {
		// At this point, we know that the user exists and is a seller, so no errors should ever occur
		logger.InternalError("Failed to count hidden seller items", "queriedUserId", queriedUserId, "error", err)
		failure_response.Unknown(context, err.Error())
		return
	}

	totalPrice, err := queries.GetSellerTotalValueOfAllItems(db, queriedUserId, queries.OnlyVisibleItems)
	if err != nil {
		// At this point, we know that the user exists and is a seller, so no errors should ever occur
		logger.InternalError("Failed to get total price of seller items", "queriedUserId", queriedUserId, "error", err)
		failure_response.Unknown(context, err.Error())
		return
	}

	response := GetSellerInformationBySellerSuccessResponse{
		ItemCount:       itemCount,
		FrozenItemCount: frozenItemCount,
		HiddenItemCount: hiddenItemCount,
		TotalPrice:      totalPrice,
	}

	context.JSON(http.StatusOK, response)
}

func (ep *GetUserInformationEndpoint) getUserInformationAsCashier(queriedUserId models.ID) {
	logger := ep.Logger
	context := ep.Context
	db := ep.Database
	userId := ep.UserID

	if userId != queriedUserId {
		logger.InvalidRequest("Cashier attempted to access another user's information", "queriedUserId", queriedUserId)
		failure_response.WrongRole(context, "Only admins can access users' information")
		return
	}

	sales, err := queries.GetSalesWithCashier(db, queriedUserId)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchUser) {
			// This should never occur. At this point we know
			// that userId == queriedUserId, and userId
			// has been checked earlier.
			logger.Bug("User not found; should have been caught earlier", "queriedUserId", queriedUserId)
			failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
			return
		}
		if errors.Is(err, dberr.ErrWrongRole) {
			// This should never occur. We previously were
			// able to determine that userId refers to a cashier,
			// and we know that userId == queriedUserId.
			logger.Bug("User has the wrong role; should have been caught earlier", "queriedUserId", queriedUserId)
			failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
			return
		}

		logger.InternalError("Failed to get sales with cashier", "queriedUserId", queriedUserId, "error", err)
		failure_response.Unknown(context, err.Error())
		return
	}

	convertedSales := algorithms.Map(sales, convertSaleToGetUserInformationSale)

	response := GetCashierInformationByCashierSuccessResponse{
		Sales: &convertedSales,
	}

	context.JSON(http.StatusOK, response)
}
