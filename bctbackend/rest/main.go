package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	rest_cashier "bctbackend/rest/cashier"
	rest_seller "bctbackend/rest/seller"
	"bctbackend/security"
	"database/sql"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

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
	router := gin.Default()
	DefineEndpoints(db, router)

	return router.Run("localhost:8000")
}

func DefineEndpoints(db *sql.DB, router *gin.Engine) {
	withUserAndRole := func(handler func(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id)) gin.HandlerFunc {
		return func(context *gin.Context) {
			sessionId, err := context.Cookie(security.SessionCookieName)

			if err != nil {
				context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: missing session ID"})
				return
			}

			sessionData, err := queries.GetSessionData(db, sessionId)

			if err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve session"})
				return
			}

			userId := sessionData.UserId
			roleId := sessionData.RoleId

			handler(context, db, userId, roleId)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	v1.POST("/login", func(context *gin.Context) { login(context, db) })
	v1.GET("/items", withUserAndRole(getItems))
	v1.GET("/sellers/:id/items", withUserAndRole(rest_seller.GetSellerItems))
	v1.POST("/sellers/:id/items", withUserAndRole(rest_seller.AddSellerItem))
	v1.GET("/sales/items/:id", withUserAndRole(rest_cashier.GetItemInformation))
	v1.POST("/sales", withUserAndRole(rest_cashier.AddSale))
}
