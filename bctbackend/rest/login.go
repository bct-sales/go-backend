package rest

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/failure_response"
	"bctbackend/security"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `form:"username" binding:"required" json:"username"`
	Password string `form:"password" binding:"required" json:"password"`
}

type LoginSuccessResponse struct {
	Role string `json:"role"`
}

// @Summary Login user.
// @Description Login user. If successful, returns the role of the user.
// @Description If the user is unknown, returns 401 Unauthorized with type "unknown_user".
// @Description If the password is wrong, returns 401 Unauthorized with type "wrong_password".
// @Success 200 {object} LoginSuccessResponse
// @Failure 400 {object} failure_response.FailureResponse "Failed to parse request"
// @Failure 401 {object} failure_response.FailureResponse "Failed to authenticate user"
// @Failure 404 {object} failure_response.FailureResponse "Unknown user"
// @Failure 500 {object} failure_response.FailureResponse "Internal error"
// @Router /login [post]
// @Param username formData string true "username"
// @Param password formData string true "password"
// @Tags authentication
func login(context *gin.Context, db *sql.DB) {
	var loginRequest LoginRequest

	if err := context.ShouldBind(&loginRequest); err != nil {
		slog.Info("Failed to parse login request", slog.String("error", err.Error()))
		failure_response.InvalidRequest(context, "Failed to parse request")
		return
	}

	userId, err := models.ParseId(loginRequest.Username)
	if err != nil {
		slog.Info("Someone tried to login with an invalid user ID", slog.String("userId", loginRequest.Username))
		failure_response.InvalidUserId(context, err.Error())
		return
	}

	password := loginRequest.Password
	roleId, err := queries.AuthenticateUser(db, userId, password)

	if err != nil {
		if errors.Is(err, database.ErrNoSuchUser) {
			slog.Info("Unknown user trying to log in", slog.String("userId", loginRequest.Username))
			failure_response.UnknownUser(context, err.Error())
			return
		}

		if errors.Is(err, database.ErrWrongPassword) {
			slog.Info("User entered wrong password", slog.String("userId", loginRequest.Username))
			failure_response.WrongPassword(context, err.Error())
			return
		}

		slog.Error("Failed authentication for unknown reasons", slog.String("userId", loginRequest.Username), slog.String("error", err.Error()))
		failure_response.Unknown(context, err.Error())
		return
	}

	expirationTime := models.Now() + security.SessionDurationInSeconds
	sessionId, err := queries.AddSession(db, userId, expirationTime)

	if err != nil {
		slog.Error("Failed to create session", slog.String("userId", loginRequest.Username), slog.String("error", err.Error()))
		failure_response.Unknown(context, err.Error())
		return
	}

	ensureSecure := false // TODO: set to true when using HTTPS
	context.SetCookie(security.SessionCookieName, string(sessionId), security.SessionDurationInSeconds, "/", "localhost", ensureSecure, true)
	roleName, err := models.NameOfRole(roleId)

	if err != nil {
		slog.Error("Failed to get role name", slog.Int64("roleId", int64(roleId)))
		failure_response.Unknown(context, err.Error())
		return
	}

	response := LoginSuccessResponse{Role: roleName}
	context.JSON(http.StatusOK, response)

	slog.Info("User logged in successfully", slog.String("userId", loginRequest.Username))
}
