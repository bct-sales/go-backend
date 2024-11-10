package seller

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AddSalePayload struct {
	Items []models.Id `json:"item_ids" binding:"required"`
}

type AddSaleResponse struct {
	SaleId models.Id `json:"sale_id"`
}

func AddSale(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.CashierRoleId {
		context.JSON(http.StatusForbidden, gin.H{"message": "Only accessible to cashiers"})
		return
	}

	var payload AddSalePayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse payload: " + err.Error()})
		return
	}

	timestamp := models.Now()

	saleId, err := queries.AddSale(
		db,
		userId,
		timestamp,
		payload.Items,
	)

	// TODO recognize error (e.g., if the category does not exist) and return StatusBadRequest or InternalServerError depending on the error
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Failed to add sale"})
		return
	}

	response := AddSaleResponse{SaleId: saleId}
	context.JSON(http.StatusCreated, response)
}
