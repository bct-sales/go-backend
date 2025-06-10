package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	rest "bctbackend/rest/shared"
	"database/sql"
	"log/slog"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type GetUsersUserData struct {
	Id           int64          `json:"id"`
	Password     string         `json:"password"`
	Role         string         `json:"role"`
	CreatedAt    rest.DateTime  `json:"createdAt"`
	LastActivity *rest.DateTime `json:"lastActivity,omitempty"`
	ItemCount    int            `json:"itemCount"`
}

type GetUsersSuccessResponse struct {
	Users []GetUsersUserData `json:"users"`
}

// @Summary Get list of users.
// @Description Returns all users. Only accessible to users with the admin role.
// @Tags users, admin
// @Accept json
// @Produce json
// @Success 200 {object} GetUsersSuccessResponse "Users successfully fetched"
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse payload or URI"
// @Failure 401 {object} failure_response.FailureResponse "Not authenticated"
// @Failure 403 {object} failure_response.FailureResponse "Only accessible to admins"
// @Failure 500 {object} failure_response.FailureResponse "Internal error"
// @Router /users [get]
func GetUsers(context *gin.Context, db *sql.DB, userId models.Id, roleId models.RoleId) {
	slog.Info(
		"User requested to fetch users",
		slog.Int64("user_id", userId.Int64()),
		slog.Int64("role_id", roleId.Int64()))

	if !roleId.IsAdmin() {
		slog.Info(
			"Non-admin attempted to list all items",
			slog.Int64("user_id", userId.Int64()),
			slog.Int64("role_id", roleId.Int64()))

		failure_response.WrongRole(context, "Only accessible to admins")
		return
	}

	users := []*queries.UserWithItemCount{}
	if err := queries.GetUsersWithItemCount(db, queries.OnlyVisibleItems, queries.CollectTo(&users)); err != nil {
		slog.Error("Failed to fetch users", slog.String("error", err.Error()))
		failure_response.Unknown(context, err.Error())
		return
	}

	var userData = []GetUsersUserData{}
	for _, user := range users {
		createdAt := rest.ConvertTimestampToDateTime(user.CreatedAt)

		var lastActivity *rest.DateTime
		if user.LastActivity == nil {
			lastActivity = nil
		} else {
			date := rest.ConvertTimestampToDateTime(*user.LastActivity)
			lastActivity = &date
		}

		userDatum := GetUsersUserData{
			Id:           user.UserId.Int64(),
			Password:     user.Password,
			Role:         user.RoleId.Name(),
			CreatedAt:    createdAt,
			LastActivity: lastActivity,
			ItemCount:    user.ItemCount,
		}

		userData = append(userData, userDatum)
	}

	response := GetUsersSuccessResponse{Users: userData}

	context.IndentedJSON(http.StatusOK, response)
}
