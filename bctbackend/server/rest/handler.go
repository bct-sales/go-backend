package rest

import (
	"bctbackend/database/models"
	"bctbackend/server/configuration"
	"bctbackend/server/logger"
	"database/sql"

	"github.com/gin-gonic/gin"
)

type HandlerFunctionArguments struct {
	Context       *gin.Context
	Configuration *configuration.Configuration
	Database      *sql.DB
	UserId        models.Id
	RoleId        models.RoleId
	Logger        logger.Logger
}

type HandlerFunction func(arguments *HandlerFunctionArguments)
