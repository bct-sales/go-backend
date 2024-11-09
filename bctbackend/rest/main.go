package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/security"
	"database/sql"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	SessionCookieName = "bctsales_session_id"
)

type LoginRequest struct {
	Username string `form:"username" binding:"required" json:"username"`
	Password string `form:"password" binding:"required" json:"password"`
}

// @Summary Login
// @Description Login
// @Success 200 {object} string
// @Router /login [post]
// @param username formData string true "username"
// @param password formData string true "password"
func login(context *gin.Context, db *sql.DB) {
	var loginRequest LoginRequest

	if err := context.ShouldBind(&loginRequest); err != nil {
		context.String(http.StatusBadRequest, "Bad request: %v", err)
		return
	}

	userId, err := models.ParseId(loginRequest.Username)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}

	password := loginRequest.Password

	if err := queries.AuthenticateUser(db, userId, password); err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Login failed"})
		return
	}

	sessionId := security.CreateSession(userId, 1, security.SessionDurationInSeconds)
	ensureSecure := false // TODO: set to true when using HTTPS
	context.SetCookie(SessionCookieName, sessionId, security.SessionDurationInSeconds, "/", "localhost", ensureSecure, true)
	context.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

// @Summary Get all items
// @Description Get all items
// @Accept json
// @Produce json
// @Success 200 {object} []models.Item
// @Router /items [get]
func getItems(context *gin.Context, db *sql.DB) {
	sessionId, err := context.Cookie(SessionCookieName)

	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: missing session ID"})
		return
	}

	session := security.GetSession(sessionId)

	if session == nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: invalid session ID"})
		return
	}

	items, err := queries.GetItems(db)

	if err != nil {
		context.AbortWithStatus(http.StatusInternalServerError)
	}

	context.IndentedJSON(http.StatusOK, items)
}

// @title           BCT Sales
// @version         1.0
// @description     BCT Sales REST API

// @contact.name   Frederic Vogels
// @contact.email  frederic.vogels@gmail.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8000
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func StartRestService(db *sql.DB) error {
	router := CreateRestRouter(db)

	return router.Run("localhost:8000")
}

func CreateRestRouter(db *sql.DB) *gin.Engine {
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	v1.POST("/login", func(context *gin.Context) { login(context, db) })
	v1.GET("/items", func(context *gin.Context) { getItems(context, db) })

	return router
}
