package rest

import (
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"bytes"
	"encoding/csv"
	"log/slog"
	"net/http"
	"strconv"

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

	users, usersOk := ep.fetchUsersWithItemCountFromDatabase()
	if !usersOk {
		return
	}

	userData := ep.convertToUserData(users)

	ep.sendSuccessResponseInAppropriateFormat(userData)
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

func (ep *listUsersEndpoint) fetchUsersWithItemCountFromDatabase() ([]*queries.UserWithItemCount, bool) {
	users := []*queries.UserWithItemCount{}

	if err := queries.GetUsersWithItemCount(ep.Database, queries.OnlyVisibleItems, queries.CollectTo(&users)); err != nil {
		ep.Logger.InternalError("Failed to fetch users", slog.String("error", err.Error()))
		failure_response.Unknown(ep.Context, err.Error())
		return nil, false
	}

	return users, true
}

func (ep *listUsersEndpoint) convertToUserData(users []*queries.UserWithItemCount) []GetUsersUserData {
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

	return userData
}

func (ep *listUsersEndpoint) sendSuccessResponseInAppropriateFormat(userData []GetUsersUserData) {
	formatHandler := formatHandlerAdapter{
		handleDefaultFormatFunc: func() { ep.sendSuccessResponseAsJSON(userData) },
		handleCSVFormatFunc:     func() { ep.sendSuccessResponseAsCSVFile(userData) },
		handleJSONFormatFunc:    func() { ep.sendSuccessResponseAsJSONFile(userData) },
	}

	ep.parseFormatQueryParameter(&formatHandler)
}

func (ep *listUsersEndpoint) sendSuccessResponseAsJSON(userData []GetUsersUserData) {
	response := GetUsersSuccessResponse{Users: userData}
	ep.Context.IndentedJSON(http.StatusOK, response)
}

func (ep *listUsersEndpoint) sendSuccessResponseAsJSONFile(userData []GetUsersUserData) {
	ep.Context.Header("Content-Type", "application/json")
	ep.Context.Header("Content-Disposition", "attachment; filename=\"users.json\"")
	ep.Context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ep.Context.Header("Pragma", "no-cache")

	ep.Context.IndentedJSON(http.StatusOK, userData)
}

func (ep *listUsersEndpoint) sendSuccessResponseAsCSVFile(userData []GetUsersUserData) {
	ep.Context.Header("Content-Type", "text/csv")
	ep.Context.Header("Content-Disposition", "attachment; filename=\"users.csv\"")
	ep.Context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ep.Context.Header("Pragma", "no-cache")

	string := ep.formatAsCSV(userData)
	if string == nil {
		return
	}

	ep.Context.String(http.StatusOK, *string)
}

func (ep *listUsersEndpoint) formatAsCSV(userData []GetUsersUserData) *string {
	buffer := new(bytes.Buffer)
	csvWriter := csv.NewWriter(buffer)

	// Write headers
	headers := []string{"user_id", "role_id", "last_activity", "password", "item_count"}
	err := csvWriter.Write(headers)
	if err != nil {
		ep.Logger.InternalError("Failed to write headers to CSV file", "error", err)
		failure_response.Unknown(ep.Context, "Failed to write headers to CSV file: "+err.Error())
		return nil
	}

	// Write rows, one per user
	for _, user := range userData {
		idString := strconv.FormatInt(user.Id, 10)
		roleString := user.Role

		var lastActivityString string
		if user.LastActivity != nil {
			lastActivityString = user.LastActivity.Timestamp.FormattedDateTime()
		} else {
			lastActivityString = "N/A"
		}

		itemCountString := strconv.FormatInt(int64(user.ItemCount), 10)

		err = csvWriter.Write([]string{
			idString,
			roleString,
			lastActivityString,
			user.Password,
			itemCountString,
		})
		if err != nil {
			ep.Logger.InternalError("Failed to write row to CSV file", "error", err)
			failure_response.Unknown(ep.Context, "Failed to write row to CSV file: "+err.Error())
			return nil
		}
	}

	csvWriter.Flush()
	result := buffer.String()
	return &result
}
