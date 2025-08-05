package rest

import (
	"bctbackend/database/models"
	"bctbackend/server/configuration"
	"database/sql"

	"github.com/gin-gonic/gin"
)

type HandlerFunctionArguments struct {
	Context       *gin.Context
	Configuration *configuration.Configuration
	Database      *sql.DB
	UserId        models.Id
	RoleId        models.RoleId
}

type HandlerFunction func(arguments *HandlerFunctionArguments)
