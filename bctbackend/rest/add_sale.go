package rest

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AddSalePayload struct {
	Items []models.Id `json:"itemIds" binding:"required"`
}

type AddSaleSuccessResponse struct {
	SaleId models.Id `json:"saleId"`
}

// @Summary Add a new sale
// @Description Adds a new sale to the database. Only accessible to users with the cashier role.
// @Tags sales
// @Accept json
// @Produce json
// @Param AddSalePayload body AddSalePayload true "Payload containing item IDs"
// @Success 201 {object} AddSaleSuccessResponse "Sale successfully added"
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to cashiers"
// @Failure 404 {object} failure_response.FailureResponse "Unknown item in sale"
// @Failure 500 {object} failure_response.FailureResponse "Internal server error"
// @Router /sales [post]
func AddSale(context *gin.Context, configuration *Configuration, db *sql.DB, userId models.Id, roleId models.RoleId) {
	if !roleId.IsCashier() {
		failure_response.WrongRole(context, "Adding sale is only accessible to cashiers")
		return
	}

	var payload AddSalePayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		slog.Error("Failed to parse AddSale payload", "error", err, "payload", payload)
		failure_response.InvalidRequest(context, "Failed to parse payload:"+err.Error())
		return
	}

	timestamp := models.Now()

	saleId, err := queries.AddSale(
		db,
		userId,
		timestamp,
		payload.Items,
	)
	if err != nil {
		if errors.Is(err, dberr.ErrSaleMissingItems) {
			failure_response.MissingItems(context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrDuplicateItemInSale) {
			failure_response.DuplicateItemInSale(context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrNoSuchItem) {
			failure_response.UnknownItem(context, err.Error())
			return
		}

		if errors.Is(err, dberr.ErrSaleRequiresCashier) {
			slog.Error("[BUG] AddSale failed with ErrSaleRequiresCashier, but this should never occur as the role is checked before", "error", err)
			failure_response.Unknown(context, "Bug: should never occur as this is checked before")
			return
		}

		slog.Error("Failed to add sale", "error", err)
		failure_response.Unknown(context, "Failed to add sale: "+err.Error())
		return
	}

	response := AddSaleSuccessResponse{SaleId: saleId}
	context.JSON(http.StatusCreated, response)
}
