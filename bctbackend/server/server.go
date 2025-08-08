package server

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/security"
	"bctbackend/server/configuration"
	"bctbackend/server/failure_response"
	"bctbackend/server/logger"
	"bctbackend/server/paths"
	"bctbackend/server/rest"
	"bctbackend/server/websocket"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

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
func StartServer(database *sql.DB, configuration *configuration.Configuration) error {
	server := NewServer(database, configuration)

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
	channel       chan int
}

func NewServer(db *sql.DB, configuration *configuration.Configuration) *Server {
	server := Server{
		database:      db,
		configuration: configuration,
		broadcaster:   websocket.NewWebsocketBroadcaster(),
		router:        createGinRouter(configuration.GinMode),
		channel:       make(chan int),
	}

	server.defineRESTEndpoints()
	server.defineWebsocketEndpoint()
	server.defineStaticFilesRoutes(configuration.HTMLPath)
	server.startPeriodicExpiredSessionCleanerUpper()

	return &server
}

func (server *Server) Shutdown() {
	slog.Info("Shutting down server")

	server.channel <- 0
	if err := server.database.Close(); err != nil {
		slog.Error("Failed to close database connection", slog.String("error", err.Error()))
	}

	slog.Info("Server shutdown complete")
}

func (server *Server) startPeriodicExpiredSessionCleanerUpper() {
	duration := time.Second
	ticker := time.NewTicker(duration)

	go func() {
		for {
			select {
			case <-ticker.C:
				now := models.Now()
				slog.Info("Cleaning up expired sessions", slog.String("current_time", now.String()))
				if err := queries.DeleteExpiredSessions(server.database, now); err != nil {
					slog.Error("Failed to clean up expired sessions", slog.String("error", err.Error()))
				}

			case <-server.channel:
				slog.Info("Stopping periodic expired session cleaner upper")
				ticker.Stop()
				return
			}
		}
	}()
}

func (server *Server) defineRESTEndpoints() {
	router := server.router

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	server.RawPOST(paths.Login(), rest.Login)
	server.RawPOST(paths.Logout(), rest.Logout)

	server.GET(paths.Items(), rest.ListAllItems)
	server.GET(paths.ItemStr(":id"), rest.GetItemInformation)
	server.PUT(paths.ItemStr(":id"), rest.UpdateItem)

	server.GET(paths.Users(), rest.ListUsers)
	server.GET(paths.UserStr(":id"), rest.GetUserInformation)

	server.GET(paths.Categories(), rest.ListCategories)

	server.GET(paths.SellerItemsStr(":id"), rest.GetSellerItems)
	server.POST(paths.SellerItemsStr(":id"), rest.AddSellerItem)

	server.POST(paths.Labels(), rest.GenerateLabels)

	server.GET(paths.Sales(), rest.GetSales)
	server.GET(paths.SaleStr(":id"), rest.GetSaleInformation)
	server.POST(paths.Sales(), rest.AddSale)
	server.GET(paths.CashierSalesStr(":id"), rest.ListCashierSales)
}

func (server *Server) defineWebsocketEndpoint() {
	server.router.GET(paths.Websocket().String(), server.broadcaster.CreateHandler())
}

func (server *Server) defineStaticFilesRoutes(htmlPath string) {
	server.router.NoRoute(func(context *gin.Context) {
		context.File(htmlPath)
	})
}

func (server *Server) RawPOST(path *paths.URL, handler func(logger logger.Logger, context *gin.Context, database *sql.DB)) {
	decoratedSlogger := slog.Default().With(slog.String("handler", getFunctionName(handler)))
	logger := logger.NewLoggerWrapper(decoratedSlogger)

	server.router.POST(path.String(), func(context *gin.Context) { handler(logger, context, server.database) })
}

func (server *Server) GET(path *paths.URL, handler rest.HandlerFunction) {
	server.router.GET(path.String(), server.withUserAndRole(handler, false))
}

func (server *Server) POST(path *paths.URL, handler rest.HandlerFunction) {
	server.router.POST(path.String(), server.withUserAndRole(handler, true))
}

func (server *Server) PUT(path *paths.URL, handler rest.HandlerFunction) {
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

	router := gin.New()

	corsConfiguration := cors.DefaultConfig()
	corsConfiguration.AllowAllOrigins = true
	// config.AllowOrigins = []string{"http://localhost:5173"}
	corsConfiguration.AllowCredentials = true
	router.Use(cors.New(corsConfiguration))

	return router
}

func (server *Server) withUserAndRole(handler rest.HandlerFunction, mutates bool) gin.HandlerFunc {
	database := server.database
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
		sessionData, err := queries.GetSessionData(database, sessionId)

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
		if err := queries.UpdateLastActivity(database, userId, now); err != nil {
			slog.Error("Failed to update last activity", slog.String("error", err.Error()))
			// Keep going, we don't want to block the request
		}

		decoratedSlogger := slog.Default().With(
			slog.String("user_id", userId.String()),
			slog.String("role_id", roleId.String()),
			slog.String("handler", getFunctionName(handler)),
		)
		logger := logger.NewLoggerWrapper(decoratedSlogger)
		arguments := rest.HandlerFunctionArguments{
			Context:       context,
			Configuration: configuration,
			Database:      database,
			UserId:        userId,
			RoleId:        roleId,
			Logger:        logger,
		}
		handler(&arguments)

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

func getFunctionName(x any) string {
	fullyQualifiedName := runtime.FuncForPC(reflect.ValueOf(x).Pointer()).Name()
	indexOfDot := strings.Index(fullyQualifiedName, ".")
	return fullyQualifiedName[indexOfDot+1:]
}
