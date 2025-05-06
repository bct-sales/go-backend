package rest

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	rest "bctbackend/rest/shared"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetUserInformationItem struct {
	ItemId       models.Id           `json:"itemId" binding:"required"`
	AddedAt      rest.DateTime       `json:"addedAt" binding:"required"`
	SellerId     models.Id           `json:"sellerId" binding:"required"`
	Description  string              `json:"description" binding:"required"`
	PriceInCents models.MoneyInCents `json:"priceInCents" binding:"required"`
	CategoryId   models.Id           `json:"categoryId" binding:"required"`
	Charity      *bool               `json:"charity" binding:"required"`
	Donation     *bool               `json:"donation" binding:"required"`
	Frozen       *bool               `json:"frozen" binding:"required"`
	SaleCount    *int                `json:"saleCount" binding:"required"`
}

type GetUserInformationSale struct {
	SaleId          models.Id        `json:"saleId" binding:"required"`
	TransactionTime models.Timestamp `json:"transactionTime" binding:"required"`
}

type GetUserInformationSuccessResponse struct {
	UserId       models.Id      `json:"userId" binding:"required"`
	Role         string         `json:"role" binding:"required"`
	Password     string         `json:"password" binding:"required"`
	CreatedAt    rest.DateTime  `json:"createdAt" binding:"required"`
	LastActivity *rest.DateTime `json:"lastActivity,omitempty"`
}

type GetSellerInformationSuccessResponse struct {
	GetUserInformationSuccessResponse
	Items *[]*GetUserInformationItem `json:"items" binding:"required"`
}

type GetAdminInformationSuccessResponse struct {
	GetUserInformationSuccessResponse
}

type GetCashierInformationSuccessResponse struct {
	GetUserInformationSuccessResponse
	Sales *[]*GetUserInformationSale `json:"sales" binding:"required"`
}

func convertItemToGetUserInformationItem(item *queries.ItemWithSaleCount) *GetUserInformationItem {
	return &GetUserInformationItem{
		ItemId:       item.ItemId,
		Description:  item.Description,
		SellerId:     item.SellerId,
		AddedAt:      rest.ConvertTimestampToDateTime(item.AddedAt),
		PriceInCents: item.PriceInCents,
		CategoryId:   item.CategoryId,
		Charity:      &item.Charity,
		Donation:     &item.Donation,
		Frozen:       &item.Frozen,
		SaleCount:    &item.SaleCount,
	}
}

func convertSaleToGetUserInformationSale(sale *models.Sale) *GetUserInformationSale {
	return &GetUserInformationSale{
		SaleId:          sale.SaleId,
		TransactionTime: sale.TransactionTime,
	}
}

// @Summary Get information about a user
// @Description Get information about a user.
// @Success 200 {object} GetItemInformationSuccessResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to admins"
// @Failure 404 {object} failure_response.FailureResponse "User not found"
// @Failure 500 {object} failure_response.FailureResponse "Internal server error"
// @Router /users/{id} [get]
func GetUserInformation(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	// Retrieve id of user whose information is being requested
	var uriParameters struct {
		UserId string `uri:"id" binding:"required"`
	}
	if err := context.ShouldBindUri(&uriParameters); err != nil {
		failure_response.InvalidUriParameters(context, "Invalid URI parameters: "+err.Error())
		return
	}

	// Parse user id
	queriedUserId, err := models.ParseId(uriParameters.UserId)
	if err != nil {
		failure_response.InvalidUserId(context, err.Error())
		return
	}

	if roleId == models.AdminRoleId {
		getUserInformationAsAdmin(context, db, queriedUserId)
		return
	} else if roleId == models.SellerRoleId {
		getUserInformationAsSeller(context, db, userId, queriedUserId)
		return
	} else if roleId == models.CashierRoleId {
		getUserInformationAsCashier(context, db, userId, queriedUserId)
		return
	} else {
		failure_response.Unknown(context, fmt.Sprintf("Bug: unhandled role %d", roleId))
		return
	}
}

func getUserInformationAsAdmin(context *gin.Context, db *sql.DB, queriedUserId models.Id) {

	// Look up user in database
	user, err := queries.GetUserWithId(db, queriedUserId)
	if err != nil {
		{
			var noSuchUserError *queries.NoSuchUserError
			if errors.As(err, &noSuchUserError) {
				failure_response.UnknownUser(context, err.Error())
				return
			}
		}

		failure_response.Unknown(context, err.Error())
		return
	}

	roleName, err := models.NameOfRole(user.RoleId)
	if err != nil {
		failure_response.Unknown(context, err.Error())
		return
	}

	basicInformation := GetUserInformationSuccessResponse{
		UserId:       user.UserId,
		Role:         roleName,
		Password:     user.Password,
		CreatedAt:    rest.ConvertTimestampToDateTime(user.CreatedAt),
		LastActivity: algorithms.MapOptional(user.LastActivity, rest.ConvertTimestampToDateTime),
	}

	if user.RoleId == models.AdminRoleId {
		response := GetAdminInformationSuccessResponse{
			GetUserInformationSuccessResponse: basicInformation,
		}
		context.JSON(http.StatusOK, response)
		return
	} else if user.RoleId == models.SellerRoleId {
		items, err := queries.GetSellerItemsWithSaleCounts(db, user.UserId)
		if err != nil {
			{
				var noSuchUserError *queries.NoSuchUserError
				if errors.As(err, &noSuchUserError) {
					failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
					return
				}
			}
			{
				var invalidRoleError *queries.InvalidRoleError
				if errors.As(err, &invalidRoleError) {
					failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
					return
				}
			}
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
	} else if user.RoleId == models.CashierRoleId {
		sales, err := queries.GetSalesWithCashier(db, user.UserId)
		if err != nil {
			{
				var noSuchUserError *queries.NoSuchUserError
				if errors.As(err, &noSuchUserError) {
					failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
					return
				}
			}
			{
				var invalidRoleError *queries.InvalidRoleError
				if errors.As(err, &invalidRoleError) {
					failure_response.Unknown(context, "Bug: should have been caught earlier. "+err.Error())
					return
				}
			}
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
		failure_response.Unknown(context, fmt.Sprintf("Bug: unhandled role %s", roleName))
		return
	}
}

func getUserInformationAsSeller(context *gin.Context, db *sql.DB, userId models.Id, queriedUserId models.Id) {
	failure_response.WrongRole(context, "Only admins can access seller information")
	return
}

func getUserInformationAsCashier(context *gin.Context, db *sql.DB, userId models.Id, queriedUserId models.Id) {
	failure_response.WrongRole(context, "Only admins can access cashier information")
	return
}
