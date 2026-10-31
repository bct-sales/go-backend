package rest

import (
	"bctbackend/database/queries"
	"bctbackend/server/failure_response"
	rest "bctbackend/server/shared"
	"bytes"
	"encoding/csv"
	"github.com/gin-gonic/gin"
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

func ListUsers(context *gin.Context, arguments *HandlerFunctionArguments) {
	endpoint := listUsersEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(context)
}

type listUsersEndpoint struct {
	Endpoint
}

func (ep *listUsersEndpoint) execute(context *gin.Context) {
	if !ep.ensureUserIsAdmin(context) {
		return
	}

	users, usersOk := ep.fetchUsersWithItemCountFromDatabase(context)
	if !usersOk {
		return
	}

	userData := ep.convertToUserData(users)

	ep.sendSuccessResponseInAppropriateFormat(context, userData)
}

func (ep *listUsersEndpoint) ensureUserIsAdmin(context *gin.Context) bool {
	logger := ep.Logger

	if !ep.RoleID.IsAdmin() {
		logger.InvalidRequest("Non-admin attempted to list all items")

		failure_response.WrongRole(context, "Only accessible to admins")

		return false
	}

	return true
}

func (ep *listUsersEndpoint) fetchUsersWithItemCountFromDatabase(context *gin.Context) ([]*queries.UserWithItemCount, bool) {
	logger := ep.Logger
	ctx := context.Request.Context()
	users := []*queries.UserWithItemCount{}

	if err := queries.GetUsersWithItemCount(ctx, ep.Database, queries.OnlyVisibleItems, queries.CollectTo(&users)); err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to fetch users")

		failure_response.Unknown(context, err.Error())

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

func (ep *listUsersEndpoint) sendSuccessResponseInAppropriateFormat(context *gin.Context, userData []GetUsersUserData) {
	formatHandler := formatHandlerAdapter{
		handleDefaultFormatFunc: func() { ep.sendSuccessResponseAsJSON(context, userData) },
		handleCSVFormatFunc:     func() { ep.sendSuccessResponseAsCSVFile(context, userData) },
		handleJSONFormatFunc:    func() { ep.sendSuccessResponseAsJSONFile(context, userData) },
	}

	ep.parseFormatQueryParameter(context, &formatHandler)
}

func (ep *listUsersEndpoint) sendSuccessResponseAsJSON(context *gin.Context, userData []GetUsersUserData) {
	response := GetUsersSuccessResponse{Users: userData}
	context.IndentedJSON(http.StatusOK, response)
}

func (ep *listUsersEndpoint) sendSuccessResponseAsJSONFile(context *gin.Context, userData []GetUsersUserData) {
	context.Header("Content-Type", "application/json")
	context.Header("Content-Disposition", "attachment; filename=\"users.json\"")
	context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	context.Header("Pragma", "no-cache")

	context.IndentedJSON(http.StatusOK, userData)
}

func (ep *listUsersEndpoint) sendSuccessResponseAsCSVFile(context *gin.Context, userData []GetUsersUserData) {
	context.Header("Content-Type", "text/csv")
	context.Header("Content-Disposition", "attachment; filename=\"users.csv\"")
	context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	context.Header("Pragma", "no-cache")

	string := ep.formatAsCSV(context, userData)
	if string == nil {
		return
	}

	context.String(http.StatusOK, *string)
}

func (ep *listUsersEndpoint) formatAsCSV(context *gin.Context, userData []GetUsersUserData) *string {
	logger := ep.Logger
	buffer := new(bytes.Buffer)
	csvWriter := csv.NewWriter(buffer)

	// Write headers
	headers := []string{"user_id", "role_id", "last_activity", "password", "item_count"}
	err := csvWriter.Write(headers)
	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to write headers to CSV file")

		failure_response.Unknown(context, "Failed to write headers to CSV file: "+err.Error())

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

			failure_response.Unknown(context, "Failed to write row to CSV file: "+err.Error())

			return nil
		}
	}

	csvWriter.Flush()
	result := buffer.String()
	return &result
}
