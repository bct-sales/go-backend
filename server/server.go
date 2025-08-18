package server

import (
	"bctbackend/clock"
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
	"io"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"

	_ "bctbackend/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type Server struct {
	loggerResources      *LoggerResources
	database             *sql.DB
	configuration        *configuration.Configuration
	broadcaster          *websocket.WebsocketBroadcaster
	router               *gin.Engine
	clock                clock.Clock
	expiredSessionTicker clock.Ticker
}

type LoggerResources struct {
	loggerFile *os.File
	logger     *slog.Logger
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
func StartServer(clock clock.Clock, database *sql.DB, configuration *configuration.Configuration) error {
	server, err := NewServer(clock, database, configuration)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	defer server.Shutdown()

	if err := server.run(); err != nil {
		return err
	}

	return nil
}

func NewServer(clock clock.Clock, db *sql.DB, configuration *configuration.Configuration) (*Server, error) {
	loggerResources, err := createLogger(configuration.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	server := Server{
		loggerResources:      loggerResources,
		database:             db,
		configuration:        configuration,
		broadcaster:          websocket.NewWebsocketBroadcaster(),
		router:               createGinRouter(configuration.Server.GinMode, loggerResources.logger),
		clock:                clock,
		expiredSessionTicker: nil,
	}

	server.defineRESTEndpoints()
	server.defineWebsocketEndpoint()
	server.defineStaticFilesRoutes(configuration.Server.HTMLPath)
	server.startPeriodicExpiredSessionPruner()

	return &server, nil
}

func createLogger(configuration *configuration.LogConfiguration) (*LoggerResources, error) {
	var writer io.Writer
	var loggerFile *os.File

	if configuration == nil {
		writer = os.Stderr
		loggerFile = nil
	} else {
		//exhaustruct:ignore
		loggerFile := lumberjack.Logger{
			Filename:   configuration.File,
			MaxSize:    configuration.MaxSizeMegabytes,
			MaxBackups: configuration.MaxBackups,
			MaxAge:     configuration.MaxAgeDays,
			Compress:   configuration.Compression,
		}

		writer = io.MultiWriter(os.Stderr, &loggerFile)
	}

	slogger := slog.New(slog.NewJSONHandler(writer, nil))
	loggerResources := LoggerResources{
		loggerFile: loggerFile,
		logger:     slogger,
	}
	return &loggerResources, nil
}

func (server *Server) Shutdown() {
	slog.Info("Shutting down server")

	server.expiredSessionTicker.Stop()

	if err := server.database.Close(); err != nil {
		slog.Error("Failed to close database connection", slog.String("error", err.Error()))
	}

	server.loggerResources.Close()

	slog.Info("Server shutdown complete")
}

func (server *Server) startPeriodicExpiredSessionPruner() {
	clock := server.clock
	pruneInterval := server.configuration.Server.ExpiredSessionPruneInterval

	if server.expiredSessionTicker != nil {
		slog.Error("Expired session ticker already exists, cannot create a new one")
		panic("bug: attempt to create multiple expired session tickers")
	}

	server.expiredSessionTicker = clock.NewTicker(pruneInterval, func() {
		now := clock.Now()
		slog.Info("Cleaning up expired sessions", slog.String("current_time", now.String()))
		if err := queries.DeleteExpiredSessions(server.database, now); err != nil {
			slog.Error("Failed to clean up expired sessions", slog.String("error", err.Error()))
		}
	})
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

func (server *Server) RawPOST(path *paths.URL, handler func(clock clock.Clock, logger logger.Logger, context *gin.Context, database *sql.DB, configuration *configuration.ServerConfiguration)) {
	decoratedSlogger := server.loggerResources.logger.With(slog.String("handler", getFunctionName(handler)))
	logger := logger.NewLoggerWrapper(decoratedSlogger)

	server.router.POST(path.String(), func(context *gin.Context) {
		handler(server.clock, logger, context, server.database, server.configuration.Server)
	})
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
	address := fmt.Sprintf("localhost:%d", server.configuration.Server.Port)

	if err := server.router.Run(address); err != nil {
		return err
	}

	return nil
}

func createGinRouter(ginMode string, logger *slog.Logger) *gin.Engine {
	gin.SetMode(ginMode)

	router := gin.New()
	router.Use(sloggin.New(logger))

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
	clock := server.clock

	return func(context *gin.Context) {
		sessionIdString, err := context.Cookie(security.SessionCookieName)
		if err != nil {
			slog.Error("Unauthorized: missing session ID")
			failure_response.MissingSessionId(context, err.Error())
			return
		}

		now := clock.Now()
		sessionId := models.SessionId(sessionIdString)
		sessionData, err := queries.GetSessionData(database, sessionId, now)

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

		if err := queries.UpdateLastActivity(database, userId, now); err != nil {
			slog.Error("Failed to update last activity", slog.String("error", err.Error()))
			// Keep going, we don't want to block the request
		}

		decoratedSlogger := server.loggerResources.logger.With(
			slog.String("user_id", userId.String()),
			slog.String("role_id", roleId.String()),
			slog.String("handler", getFunctionName(handler)),
		)
		logger := logger.NewLoggerWrapper(decoratedSlogger)
		arguments := rest.HandlerFunctionArguments{
			Clock:         clock,
			Context:       context,
			Configuration: configuration,
			Database:      rest.NewDatabaseWrapper(context, database),
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

func (l *LoggerResources) Close() {
	l.loggerFile.Close()
}
