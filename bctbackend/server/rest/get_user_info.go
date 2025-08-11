package rest

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	"bctbackend/server/logger"
	rest "bctbackend/server/shared"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetUserInformationItem struct {
	ItemId       models.Id           `binding:"required" json:"itemId"`
	AddedAt      rest.DateTime       `binding:"required" json:"addedAt"`
	SellerId     models.Id           `binding:"required" json:"sellerId"`
	Description  string              `binding:"required" json:"description"`
	PriceInCents models.MoneyInCents `binding:"required" json:"priceInCents"`
	CategoryId   models.Id           `binding:"required" json:"categoryId"`
	Charity      *bool               `binding:"required" json:"charity"`
	Donation     *bool               `binding:"required" json:"donation"`
	Frozen       *bool               `binding:"required" json:"frozen"`
	SaleCount    *int                `binding:"required" json:"saleCount"`
}

type GetUserInformationSale struct {
	SaleId          models.Id     `binding:"required" json:"saleId"`
	TransactionTime rest.DateTime `binding:"required" json:"transactionTime"`
}

type GetUserInformationSuccessResponse struct {
	UserId       models.Id      `binding:"required"            json:"userId"`
	Role         string         `binding:"required"            json:"role"`
	Password     string         `binding:"required"            json:"password"`
	CreatedAt    rest.DateTime  `binding:"required"            json:"createdAt"`
	LastActivity *rest.DateTime `json:"lastActivity,omitempty"`
}

type GetSellerInformationSuccessResponse struct {
	GetUserInformationSuccessResponse
	Items *[]*GetUserInformationItem `binding:"required" json:"items"`
}

type GetAdminInformationSuccessResponse struct {
	GetUserInformationSuccessResponse
}

type GetCashierInformationSuccessResponse struct {
	GetUserInformationSuccessResponse
	Sales *[]*GetUserInformationSale `binding:"required" json:"sales"`
}

type GetSellerSummarySuccessResponse struct {
	ItemCount       int                 `binding:"required" json:"itemCount"`
	FrozenItemCount int                 `binding:"required" json:"frozenItemCount"`
	HiddenItemCount int                 `binding:"required" json:"hiddenItemCount"`
	TotalPrice      models.MoneyInCents `binding:"required" json:"totalPrice"`
}

func convertItemToGetUserInformationItem(item *queries.ItemWithSaleCount) *GetUserInformationItem {
	return &GetUserInformationItem{
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

func convertSaleToGetUserInformationSale(sale *models.Sale) *GetUserInformationSale {
	return &GetUserInformationSale{
		SaleId:          sale.SaleID,
		TransactionTime: rest.ConvertTimestampToDateTime(sale.TransactionTime),
	}
}

// @Summary Get information about a user
// @Description Get information about a user.
// @Success 200 {object} GetSellerSummarySuccessResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to admins"
// @Failure 404 {object} failure_response.FailureResponse "User not found"
// @Failure 500 {object} failure_response.FailureResponse "Internal server error"
// @Router /users/{id} [get]
func GetUserInformation(arguments *HandlerFunctionArguments) {
	context := arguments.Context
	userId := arguments.UserId
	roleId := arguments.RoleId
	db := arguments.Database
	logger := arguments.Logger

	// Retrieve id of user whose information is being requested
	var uriParameters struct {
		UserId string `binding:"required" uri:"id"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		logger.InvalidInput("Invalid URI parameters", "error", err)
		failure_response.InvalidUriParameters(context, "Invalid URI parameters: "+err.Error())
		return
	}

	// Parse user id
	queriedUserId, err := models.ParseId(uriParameters.UserId)
	if err != nil {
		logger.InvalidInput("Invalid user ID", "error", err, "userId", uriParameters.UserId)
		failure_response.InvalidUserId(context, err.Error())
		return
	}

	if roleId.IsAdmin() {
		// If the user is an admin, they can access any user's information
		getUserInformationAsAdmin(logger, context, db, queriedUserId)
		return
	} else if roleId.IsSeller() {
		getUserInformationAsSeller(logger, context, db, userId, queriedUserId)
		return
	} else if roleId.IsCashier() {
		getUserInformationAsCashier(logger, context, db, userId, queriedUserId)
		return
	} else {
		failure_response.Unknown(context, fmt.Sprintf("Bug: unhandled role %d", roleId))
		return
	}
}

func getUserInformationAsAdmin(logger logger.Logger, context *gin.Context, db *sql.DB, queriedUserId models.Id) {
	// Look up user in database
	user, err := queries.GetUserWithId(db, queriedUserId)
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

	basicInformation := GetUserInformationSuccessResponse{
		UserId:       user.UserId,
		Role:         user.RoleId.Name(),
		Password:     user.Password,
		CreatedAt:    rest.ConvertTimestampToDateTime(user.CreatedAt),
		LastActivity: algorithms.MapOptional(user.LastActivity, rest.ConvertTimestampToDateTime),
	}

	if user.RoleId.IsAdmin() {
		response := GetAdminInformationSuccessResponse{
			GetUserInformationSuccessResponse: basicInformation,
		}
		context.JSON(http.StatusOK, response)
		return
	} else if user.RoleId.IsSeller() {
		items, err := queries.GetSellerItemsWithSaleCounts(db, user.UserId)
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

		response := GetSellerInformationSuccessResponse{
			GetUserInformationSuccessResponse: basicInformation,
			Items:                             &convertedItems,
		}

		context.JSON(http.StatusOK, response)
		return
	} else if user.RoleId.IsCashier() {
		sales, err := queries.GetSalesWithCashier(db, user.UserId)
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

		response := GetCashierInformationSuccessResponse{
			GetUserInformationSuccessResponse: basicInformation,
			Sales:                             &convertedSales,
		}
		context.JSON(http.StatusOK, response)
		return
	} else {
		logger.Bug("Unhandled user role")
		failure_response.Unknown(context, fmt.Sprintf("Bug: unhandled role %d", user.RoleId.Int64()))
		return
	}
}

func getUserInformationAsSeller(logger logger.Logger, context *gin.Context, db *sql.DB, userId models.Id, queriedUserId models.Id) {
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

	response := GetSellerSummarySuccessResponse{
		ItemCount:       itemCount,
		FrozenItemCount: frozenItemCount,
		HiddenItemCount: hiddenItemCount,
		TotalPrice:      models.MoneyInCents(totalPrice),
	}

	context.JSON(http.StatusOK, response)
}

func getUserInformationAsCashier(logger logger.Logger, context *gin.Context, db *sql.DB, userId models.Id, queriedUserId models.Id) {
	if userId != queriedUserId {
		logger.InvalidRequest("Cashier attempted to access another user's information", "queriedUserId", queriedUserId)
		failure_response.WrongRole(context, "Only admins can access users' information")
		return
	}

	failure_response.Forbidden(context, "not_yet_implemented", "Cashiers have no information (as of yet)")
}
