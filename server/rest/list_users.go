package rest

import (
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"log/slog"
	"net/http"

	_ "bctbackend/docs"
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
func ListUsers(arguments *HandlerFunctionArguments) {
	endpoint := listUsersEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

type listUsersEndpoint struct {
	Endpoint
}

func (ep *listUsersEndpoint) execute() {
	if !ep.ensureUserIsAdmin() {
		return
	}

	

	users := []*queries.UserWithItemCount{}
	if err := queries.GetUsersWithItemCount(ep.Database, queries.OnlyVisibleItems, queries.CollectTo(&users)); err != nil {
		ep.Logger.InternalError("Failed to fetch users", slog.String("error", err.Error()))
		failure_response.Unknown(ep.Context, err.Error())
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

	ep.Context.IndentedJSON(http.StatusOK, response)
}

func (ep *listUsersEndpoint) ensureUserIsAdmin() bool {
	if !ep.RoleId.IsAdmin() {
		ep.Logger.InvalidRequest(
			"Non-admin attempted to list all items",
			slog.Int64("user_id", ep.UserId.Int64()),
			slog.Int64("role_id", ep.RoleId.Int64()))

		failure_response.WrongRole(ep.Context, "Only accessible to admins")

		return false
	}

	return true
}
