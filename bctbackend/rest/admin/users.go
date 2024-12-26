package admin

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type GetUsersFailureResponse struct {
	Message string `json:"message"`
}

// @Summary Get list of users.
// @Description Returns all users. Only accessible to users with the admin role.
// @Tags users, admin
// @Accept json
// @Produce json
// @Success 200 {object} []models.User "Users successfully fetched"
// @Failure 403 {object} GetItemsFailureResponse "Unauthorized access"
// @Failure 500 {object} GetItemsFailureResponse "Failed to fetch items"
// @Router /users [get]
func GetUsers(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.AdminRoleId {
		failureResponse := GetItemsFailureResponse{Message: "Only accessible to admins"}
		context.JSON(http.StatusForbidden, failureResponse)
		return
	}

	users, err := queries.ListUsers(db)

	if err != nil {
		failureResponse := GetUsersFailureResponse{Message: "Failed to fetch users"}
		context.JSON(http.StatusInternalServerError, failureResponse)
		return
	}

	context.IndentedJSON(http.StatusOK, users)
}
