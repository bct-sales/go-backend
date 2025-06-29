package server

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/security"
	"bctbackend/server/configuration"
	"bctbackend/server/failure_response"
	"bctbackend/server/paths"
	"bctbackend/server/rest"
	"bctbackend/server/websocket"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	_ "bctbackend/docs"

	"github.com/gin-contrib/cors"
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
func StartServer(db *sql.DB, configuration *configuration.Configuration) error {
	server := NewServer(db, configuration)

	if err := server.run(); err != nil {
		return err
	}

	return nil
}

type Server struct {
	database      *sql.DB
	configuration *configuration.Configuration
	broadcaster   *websocket.WebsocketBroadcaster
	router        *gin.Engine
}

func NewServer(db *sql.DB, configuration *configuration.Configuration) *Server {
	server := Server{
		database:      db,
		configuration: configuration,
		broadcaster:   websocket.NewWebsocketBroadcaster(),
		router:        createGinRouter(configuration.GinMode),
	}

	server.defineEndpoints()

	return &server
}

func (server *Server) defineEndpoints() {
	router := server.router

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	server.RawPOST(paths.Login(), rest.Login)
	server.RawPOST(paths.Logout(), rest.Logout)

	server.GET(paths.Items(), rest.GetAllItems)
	server.GET(paths.ItemStr(":id"), rest.GetItemInformation)
	server.PUT(paths.ItemStr(":id"), rest.UpdateItem)

	server.GET(paths.Users(), rest.GetUsers)
	server.GET(paths.UserStr(":id"), rest.GetUserInformation)

	server.GET(paths.Categories(), rest.ListCategories)

	server.GET(paths.SellerItemsStr(":id"), rest.GetSellerItems)
	server.POST(paths.SellerItemsStr(":id"), rest.AddSellerItem)

	server.POST(paths.Labels(), rest.GenerateLabels)

	server.GET(paths.Sales(), rest.GetSales)
	server.GET(paths.SaleStr(":id"), rest.GetSaleInformation)
	server.POST(paths.Sales(), rest.AddSale)
	server.GET(paths.CashierSalesStr(":id"), rest.GetCashierSales)

	router.GET(paths.Websocket().String(), server.broadcaster.CreateHandler())
}

func (server *Server) RawPOST(path *paths.URL, handler func(context *gin.Context, database *sql.DB)) {
	server.router.POST(path.String(), func(context *gin.Context) { handler(context, server.database) })
}

func (server *Server) GET(path *paths.URL, handler HandlerFunction) {
	server.router.GET(path.String(), server.withUserAndRole(handler, false))
}

func (server *Server) POST(path *paths.URL, handler HandlerFunction) {
	server.router.POST(path.String(), server.withUserAndRole(handler, true))
}

func (server *Server) PUT(path *paths.URL, handler HandlerFunction) {
	server.router.PUT(path.String(), server.withUserAndRole(handler, true))
}

func (server *Server) run() error {
	address := fmt.Sprintf("localhost:%d", server.configuration.Port)

	if err := server.router.Run(address); err != nil {
		return err
	}

	return nil
}

func createGinRouter(ginMode string) *gin.Engine {
	gin.SetMode(ginMode)

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	// config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowCredentials = true

	router.Use(cors.New(config))

	return router
}

type HandlerFunction func(context *gin.Context, configuration *configuration.Configuration, db *sql.DB, userId models.Id, roleId models.RoleId)

func (server *Server) withUserAndRole(handler HandlerFunction, mutates bool) gin.HandlerFunc {
	db := server.database
	configuration := server.configuration
	broadcaster := server.broadcaster

	return func(context *gin.Context) {
		sessionIdString, err := context.Cookie(security.SessionCookieName)
		if err != nil {
			slog.Error("Unauthorized: missing session ID")
			failure_response.MissingSessionId(context, err.Error())
			return
		}

		sessionId := models.SessionId(sessionIdString)
		sessionData, err := queries.GetSessionData(db, sessionId)

		if errors.Is(err, dberr.ErrNoSuchSession) {
			slog.Error("Session not found")
			failure_response.NoSuchSession(context, err.Error())
			return
		}

		if err != nil {
			slog.Error("Failed to retrieve session from database", slog.String("error", err.Error()))
			failure_response.Unknown(context, "Failed to retrieve session from database: "+err.Error())
			return
		}

		userId := sessionData.UserId
		roleId := sessionData.RoleId

		now := models.Now()
		if err := queries.UpdateLastActivity(db, userId, now); err != nil {
			slog.Error("Failed to update last activity", slog.String("error", err.Error()))
			// Keep going, we don't want to block the request
		}

		handler(context, configuration, db, userId, roleId)

		if mutates {
			broadcaster.Broadcast("update")
		}
	}
}

func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if server.router == nil {
		panic("Server router is not initialized")
	}

	server.router.ServeHTTP(writer, request)
}
