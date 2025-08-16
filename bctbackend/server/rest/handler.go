package rest

import (
	"bctbackend/clock"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/configuration"
	"bctbackend/server/logger"
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
)

type HandlerFunctionArguments struct {
	Clock         clock.Clock
	Context       *gin.Context
	Configuration *configuration.Configuration
	Database      Database
	UserId        models.Id
	RoleId        models.RoleId
	Logger        logger.Logger
}

type HandlerFunction func(arguments *HandlerFunctionArguments)

type Database interface {
	queries.DatabaseQuerier

	StartTransaction() (*queries.TransactionalDatabaseQuerier, error)
}

type DatabaseWrapper struct {
	queries.ContextDatabaseQuerier
}

func NewDatabaseWrapper(ctx context.Context, db *sql.DB) *DatabaseWrapper {
	return &DatabaseWrapper{
		ContextDatabaseQuerier: *queries.NewContextDatabaseQuerier(db, ctx),
	}
}

func (w *DatabaseWrapper) StartTransaction() (*queries.TransactionalDatabaseQuerier, error) {
	return queries.NewTransactionDatabaseQuerier(w.Context, w.Database)
}
