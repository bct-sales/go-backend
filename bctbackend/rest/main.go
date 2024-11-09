package rest

import (
	"bctbackend/database/queries"
	"bctbackend/security"
	"database/sql"
	"net/http"
	"strconv"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	SessionCookieName = "bctsales_session_id"
)

func login(context *gin.Context, db *sql.DB) {
	userIdAsString := context.PostForm("username")
	password := context.PostForm("password")

	userId, err := strconv.ParseInt(userIdAsString, 10, 64)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := queries.AuthenticateUser(db, userId, password); err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Login failed"})
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
