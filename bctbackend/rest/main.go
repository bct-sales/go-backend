package rest

import (
	"bctbackend/database/queries"
	"database/sql"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

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
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	v1.GET("/items", func(context *gin.Context) { getItems(context, db) })

	return router.Run("localhost:8000")
}
