package rest

import (
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"bytes"
	"encoding/csv"
	"net/http"
	"strconv"
)

type GetUsersUserData struct {
	ID           int64          `json:"id"`
	Password     string         `json:"password"`
	Role         string         `json:"role"`
	CreatedAt    rest.DateTime  `json:"createdAt"`
	LastActivity *rest.DateTime `json:"lastActivity,omitempty"`
	ItemCount    int            `json:"itemCount"`
}

type GetUsersSuccessResponse struct {
	Users []GetUsersUserData `json:"users"`
}

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
	logger := ep.Logger

	if !ep.RoleID.IsAdmin() {
		logger.InvalidRequest("Non-admin attempted to list all items")

		failure_response.WrongRole(ep.Context, "Only accessible to admins")

		return false
	}

	return true
}

func (ep *listUsersEndpoint) fetchUsersWithItemCountFromDatabase() ([]*queries.UserWithItemCount, bool) {
	logger := ep.Logger
	users := []*queries.UserWithItemCount{}

	if err := queries.GetUsersWithItemCount(ep.Database, queries.OnlyVisibleItems, queries.CollectTo(&users)); err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to fetch users")

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
			ID:           user.UserID.Int64(),
			Password:     user.Password,
			Role:         user.RoleID.Name(),
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
	logger := ep.Logger
	buffer := new(bytes.Buffer)
	csvWriter := csv.NewWriter(buffer)

	// Write headers
	headers := []string{"user_id", "role_id", "last_activity", "password", "item_count"}
	err := csvWriter.Write(headers)
	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to write headers to CSV file")

		failure_response.Unknown(ep.Context, "Failed to write headers to CSV file: "+err.Error())

		return nil
	}

	// Write rows, one per user
	for _, user := range userData {
		idString := strconv.FormatInt(user.ID, 10)
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
			logger.AddInformation("error", err)
			logger.InternalError("Failed to write row to CSV file")

			failure_response.Unknown(ep.Context, "Failed to write row to CSV file: "+err.Error())

			return nil
		}
	}

	csvWriter.Flush()
	result := buffer.String()
	return &result
}
