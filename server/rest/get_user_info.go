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
	UserID       models.ID      `binding:"required"            json:"userId"`
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
	ItemID       models.ID           `binding:"required" json:"itemId"`
	AddedAt      rest.DateTime       `binding:"required" json:"addedAt"`
	SellerID     models.ID           `binding:"required" json:"sellerId"`
	Description  string              `binding:"required" json:"description"`
	PriceInCents models.MoneyInCents `binding:"required" json:"priceInCents"`
	CategoryID   models.ID           `binding:"required" json:"categoryId"`
	Charity      *bool               `binding:"required" json:"charity"`
	Donation     *bool               `binding:"required" json:"donation"`
	Large        *bool               `binding:"required" json:"large"`
	Frozen       *bool               `binding:"required" json:"frozen"`
	SaleCount    *int                `binding:"required" json:"saleCount"`
}

// GetCashierInformationByAdminSaleData contains data regarding sales.
// It is used when an admin requests information about a cashier.
type GetCashierInformationByAdminSaleData struct {
	SaleID          models.ID     `binding:"required" json:"saleId"`
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
		ItemID:       item.ItemID,
		Description:  item.Description,
		SellerID:     item.SellerID,
		AddedAt:      rest.ConvertTimestampToDateTime(item.AddedAt),
		PriceInCents: item.PriceInCents,
		CategoryID:   item.CategoryID,
		Charity:      &item.Charity,
		Donation:     &item.Donation,
		Large:        &item.Large,
		Frozen:       &item.Frozen,
		SaleCount:    &item.SaleCount,
	}
}

func convertSaleToGetUserInformationSale(sale *models.Sale) *GetCashierInformationByAdminSaleData {
	return &GetCashierInformationByAdminSaleData{
		SaleID:          sale.SaleID,
		TransactionTime: rest.ConvertTimestampToDateTime(sale.TransactionTime),
	}
}

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
	roleID := ep.RoleID

	queriedUserID, err := ep.retrieveQueriedUserFromUri()
	if err != nil {
		return
	}

	if roleID.IsAdmin() {
		// If the user is an admin, they can access any user's information
		ep.getUserInformationAsAdmin(queriedUserID)
		return
	} else if roleID.IsSeller() {
		ep.getUserInformationAsSeller(queriedUserID)
		return
	} else if roleID.IsCashier() {
		ep.getUserInformationAsCashier(queriedUserID)
		return
	} else {
		failure_response.Unknown(context, fmt.Sprintf("Bug: unhandled role %d", roleID))
		return
	}
}

func (ep *GetUserInformationEndpoint) retrieveQueriedUserFromUri() (models.ID, error) {
	logger := ep.Logger

	// Retrieve id of user whose information is being requested
	var uriParameters struct {
		UserID string `binding:"required" uri:"id"`
	}
	if err := ep.Context.ShouldBindUri(&uriParameters); err != nil {
		logger.AddInformation("error", err)
		logger.InvalidInput("Invalid URI parameters")

		failure_response.InvalidUriParameters(ep.Context, "Invalid URI parameters: "+err.Error())
		return 0, err
	}

	// Parse user id
	queriedUserID, err := models.ParseID(uriParameters.UserID)
	if err != nil {
		logger.AddInformation("error", err)
		logger.AddInformation("userID", uriParameters.UserID)
		logger.InvalidInput("Invalid user ID")

		failure_response.InvalidUserID(ep.Context, err.Error())

		return 0, err
	}

	logger.AddInformation("queriedUserID", queriedUserID)

	return queriedUserID, nil
}

func (ep *GetUserInformationEndpoint) getUserInformationAsAdmin(queriedUserID models.ID) {
	logger := ep.Logger
	context := ep.Context
	db := ep.Database

	// Look up user in database
	user, err := queries.GetUserWithID(db, queriedUserID)
	if err != nil {
		logger.AddInformation("error", err)

		if errors.Is(err, dberr.ErrNoSuchUser) {
			logger.InvalidRequest("User not found")

			failure_response.UnknownUser(context, err.Error())

			return
		}

		logger.InternalError("Could not find user in database")
		failure_response.Unknown(context, err.Error())
		return
	}

	basicInformation := GetUserInformationByAdminSuccessResponse{
		UserID:       user.UserID,
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
			logger.AddInformation("error", err)

			{
				if errors.Is(err, dberr.ErrNoSuchUser) {
					logger.Bug("User not found; should have been caught earlier")

					failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())

					return
				}
			}
			if errors.Is(err, dberr.ErrWrongRole) {
				logger.Bug("User has the wrong role; should have been caught earlier")
				failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
				return
			}

			logger.InternalError("Failed to get seller items with sale counts")
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
			logger.AddInformation("error", err)

			if errors.Is(err, dberr.ErrNoSuchUser) {
				logger.Bug("User not found; should have been caught earlier")

				failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())

				return
			}
			if errors.Is(err, dberr.ErrWrongRole) {
				logger.Bug("User has the wrong role; should have been caught earlier")

				failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())

				return
			}

			logger.InternalError("Failed to get sales with cashier")

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

func (ep *GetUserInformationEndpoint) getUserInformationAsSeller(queriedUserID models.ID) {
	logger := ep.Logger
	context := ep.Context
	db := ep.Database
	userID := ep.UserID

	if userID != queriedUserID {
		logger.InvalidRequest("Seller attempted to access another user's information")
		failure_response.WrongRole(context, "Only admins can access other users' information")
		return
	}

	itemCount, err := queries.CountSellerItems(db, queriedUserID, queries.IncludeAll, queries.Exclude)
	if err != nil {
		logger.AddInformation("error", err)
		// At this point, we know that the user exists and is a seller, so no errors should ever occur
		logger.InternalError("Failed to count seller items")

		failure_response.Unknown(context, err.Error())

		return
	}

	frozenItemCount, err := queries.CountSellerItems(db, queriedUserID, queries.IncludeOnly, queries.IncludeAll)
	if err != nil {
		// At this point, we know that the user exists and is a seller, so no errors should ever occur
		logger.AddInformation("error", err)
		logger.InternalError("Failed to count frozen seller items")

		failure_response.Unknown(context, err.Error())

		return
	}

	hiddenItemCount, err := queries.CountSellerItems(db, queriedUserID, queries.IncludeAll, queries.IncludeOnly)
	if err != nil {
		// At this point, we know that the user exists and is a seller, so no errors should ever occur
		logger.AddInformation("error", err)
		logger.InternalError("Failed to count hidden seller items")

		failure_response.Unknown(context, err.Error())

		return
	}

	totalPrice, err := queries.GetSellerTotalValueOfAllItems(db, queriedUserID, queries.OnlyVisibleItems)
	if err != nil {
		// At this point, we know that the user exists and is a seller, so no errors should ever occur
		logger.AddInformation("error", err)
		logger.InternalError("Failed to get total price of seller items")

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

func (ep *GetUserInformationEndpoint) getUserInformationAsCashier(queriedUserID models.ID) {
	logger := ep.Logger
	context := ep.Context
	db := ep.Database
	userID := ep.UserID

	if userID != queriedUserID {
		logger.InvalidRequest("Cashier attempted to access another user's information")

		failure_response.WrongRole(context, "Only admins can access users' information")

		return
	}

	sales, err := queries.GetSalesWithCashier(db, queriedUserID)
	if err != nil {
		logger.AddInformation("error", err)

		if errors.Is(err, dberr.ErrNoSuchUser) {
			// This should never occur. At this point we know
			// that userID == queriedUserID, and userID
			// has been checked earlier.
			logger.Bug("User not found; should have been caught earlier")

			failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())

			return
		}
		if errors.Is(err, dberr.ErrWrongRole) {
			// This should never occur. We previously were
			// able to determine that userID refers to a cashier,
			// and we know that userID == queriedUserID.
			logger.Bug("User has the wrong role; should have been caught earlier")

			failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())

			return
		}

		logger.InternalError("Failed to get sales with cashier")

		failure_response.Unknown(context, err.Error())

		return
	}

	convertedSales := algorithms.Map(sales, convertSaleToGetUserInformationSale)

	response := GetCashierInformationByCashierSuccessResponse{
		Sales: &convertedSales,
	}

	context.JSON(http.StatusOK, response)
}
