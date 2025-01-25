package admin

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	rest "bctbackend/rest/shared"
	"database/sql"
	"log/slog"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type GetUsersUserData struct {
	Id        int64                    `json:"id"`
	Password  string                   `json:"password"`
	Role      string                   `json:"role"`
	CreatedAt rest.StructuredTimestamp `json:"created_at"`
}

type GetUsersSuccessResponse struct {
	Users []GetUsersUserData `json:"users"`
}

type GetUsersFailureResponse struct {
	Message string `json:"message"`
}

// @Summary Get list of users.
// @Description Returns all users. Only accessible to users with the admin role.
// @Tags users, admin
// @Accept json
// @Produce json
// @Success 200 {object} GetUsersSuccessResponse "Users successfully fetched"
// @Failure 403 {object} GetItemsFailureResponse "Unauthorized access"
// @Failure 500 {object} GetItemsFailureResponse "Internal error"
// @Router /users [get]
func GetUsers(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	slog.Info("User requested to fetch users", slog.Int64("user_id", userId), slog.Int64("role_id", roleId))

	if roleId != models.AdminRoleId {
		slog.Info("User attempted to access /users but is not an admin", slog.Int64("user_id", userId), slog.Int64("role_id", roleId))
		failureResponse := GetUsersFailureResponse{Message: "Only accessible to admins"}
		context.JSON(http.StatusForbidden, failureResponse)
		return
	}

	users, err := queries.GetUsers(db)

	if err != nil {
		slog.Error("Failed to fetch users", slog.String("error", err.Error()))
		failureResponse := GetUsersFailureResponse{Message: "Failed to fetch users"}
		context.JSON(http.StatusInternalServerError, failureResponse)
		return
	}

	var userData []GetUsersUserData = []GetUsersUserData{}

	for _, user := range users {
		roleName, err := models.NameOfRole(user.RoleId)

		if err != nil {
			slog.Error("Failed to translate role ID to role name", slog.String("error", err.Error()))
			failureResponse := GetUsersFailureResponse{Message: "Failed to translate role ID to role name"}
			context.JSON(http.StatusInternalServerError, failureResponse)
			return
		}

		createdAt := rest.FromUnix(user.CreatedAt)

		userDatum := GetUsersUserData{
			Id:        user.UserId,
			Password:  user.Password,
			Role:      roleName,
			CreatedAt: createdAt,
		}

		userData = append(userData, userDatum)
	}

	response := GetUsersSuccessResponse{Users: userData}

	context.IndentedJSON(http.StatusOK, response)
}
