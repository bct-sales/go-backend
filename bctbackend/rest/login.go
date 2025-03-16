package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
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

type LoginFailureResponse struct {
	Type    string `json:"type"`
	Details string `json:"details"`
}

const (
	LoginFailureType_BadRequest    = "bad_request"
	LoginFailureType_InvalidUserId = "invalid_id"
	LoginFailureType_UnknownUser   = "unknown_user"
	LoginFailureType_WrongPassword = "wrong_password"
	LoginFailureType_Unknown       = "unknown"
)

// @Summary Login user
// @Description Login user. If successful, returns the role of the user.
// @Description If the user is unknown, returns 401 Unauthorized with type "unknown_user".
// @Description If the password is wrong, returns 401 Unauthorized with type "wrong_password".
// @Success 200 {object} LoginSuccessResponse
// @Failure 400 {object} LoginFailureResponse "Failed to parse request"
// @Failure 401 {object} LoginFailureResponse "Failure to authenticate user"
// @Failure 500 {object} LoginFailureResponse "Internal error"
// @Router /login [post]
// @Param username formData string true "username"
// @Param password formData string true "password"
// @Tags authentication
func login(context *gin.Context, db *sql.DB) {
	var loginRequest LoginRequest

	if err := context.ShouldBind(&loginRequest); err != nil {
		slog.Info("Failed to parse login request", slog.String("error", err.Error()))
		failureResponse := LoginFailureResponse{Type: LoginFailureType_BadRequest, Details: err.Error()}
		context.JSON(http.StatusBadRequest, failureResponse)
		return
	}

	userId, err := models.ParseId(loginRequest.Username)

	if err != nil {
		slog.Info("Someone tried to login with invalid user ID", slog.String("userId", loginRequest.Username))
		failureResponse := LoginFailureResponse{Type: LoginFailureType_InvalidUserId, Details: err.Error()}
		context.JSON(http.StatusUnauthorized, failureResponse)
		return
	}

	password := loginRequest.Password
	roleId, err := queries.AuthenticateUser(db, userId, password)

	if err != nil {
		var noSuchUserError *queries.NoSuchUserError
		if errors.As(err, &noSuchUserError) {
			slog.Info("Unknown user trying to log in", slog.String("userId", loginRequest.Username))
			failureResponse := LoginFailureResponse{Type: LoginFailureType_UnknownUser, Details: err.Error()}
			context.JSON(http.StatusUnauthorized, failureResponse)
			return
		}

		var wrongPasswordError *queries.WrongPasswordError
		if errors.As(err, &wrongPasswordError) {
			slog.Info("User entered wrong password", slog.String("userId", loginRequest.Username))
			failureResponse := LoginFailureResponse{Type: LoginFailureType_WrongPassword, Details: err.Error()}
			context.JSON(http.StatusUnauthorized, failureResponse)
			return
		}

		slog.Error("Failed authentication for unknown reasons", slog.String("userId", loginRequest.Username), slog.String("error", err.Error()))
		failureResponse := LoginFailureResponse{Type: LoginFailureType_Unknown, Details: err.Error()}
		context.JSON(http.StatusUnauthorized, failureResponse)
		return
	}

	expirationTime := models.Now() + security.SessionDurationInSeconds
	sessionId, err := queries.AddSession(db, userId, expirationTime)

	if err != nil {
		slog.Error("Failed to create session", slog.String("userId", loginRequest.Username), slog.String("error", err.Error()))
		failureResponse := LoginFailureResponse{Type: LoginFailureType_Unknown, Details: err.Error()}
		context.AbortWithStatusJSON(http.StatusInternalServerError, failureResponse)
		return
	}

	ensureSecure := false // TODO: set to true when using HTTPS
	context.SetCookie(security.SessionCookieName, sessionId, security.SessionDurationInSeconds, "/", "localhost", ensureSecure, true)
	roleName, err := models.NameOfRole(roleId)

	if err != nil {
		slog.Error("Failed to get role name", slog.Int64("roleId", int64(roleId)))
		failureResponse := LoginFailureResponse{Type: LoginFailureType_Unknown, Details: err.Error()}
		context.JSON(http.StatusInternalServerError, failureResponse)
		return
	}

	response := LoginSuccessResponse{Role: roleName}
	context.JSON(http.StatusOK, response)

	slog.Info("User logged in successfully", slog.String("userId", loginRequest.Username))
}
