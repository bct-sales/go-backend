package rest

import (
	"bctbackend/clock"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server/configuration"
	"bctbackend/server/failure_response"
	"bctbackend/server/logger"
	"context"
	"database/sql"
	"strconv"

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
		ContextDatabaseQuerier: queries.ContextDatabaseQuerier{
			Database: db,
			Context:  ctx,
		},
	}
}

func (w *DatabaseWrapper) StartTransaction() (*queries.TransactionalDatabaseQuerier, error) {
	return queries.NewTransactionDatabaseQuerier(w.Context, w.Database)
}

type Endpoint struct {
	HandlerFunctionArguments
}

// parseOffsetQueryParameter looks for a query parameter named offset.
// If it is missing, nil, true is returned.
// If it is present and its value is a valid integer, the parsed integer and true are returned.
// If it is present and its value is invalid, nil, false is returned.
func (ep *Endpoint) parseOffsetQueryParameter() (*int, bool) {
	offsetString := ep.Context.Query("offset")
	if offsetString != "" {
		parsedOffset, err := strconv.Atoi(offsetString)

		if err != nil {
			ep.Logger.InvalidInput("Failed to parse offset", "error", err, "offset", offsetString)
			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Failed to parse offset: "+err.Error())
			return nil, false
		}

		if parsedOffset < 0 {
			ep.Logger.InvalidRequest("Invalid offset parameter", "offset", offsetString)
			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Offset must be 0 or greater")
			return nil, false
		}

		return &parsedOffset, true
	} else {
		return nil, true
	}
}

func (ep *Endpoint) parseLimitQueryParameter() (*int, bool) {
	limitString := ep.Context.Query("limit")
	if limitString != "" {
		parsedLimit, err := strconv.Atoi(limitString)

		if err != nil {
			ep.Logger.InvalidInput("Failed to parse limit", "error", err, "limit", limitString)
			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Failed to parse limit: "+err.Error())
			return nil, false
		}

		if parsedLimit < 1 {
			ep.Logger.InvalidRequest("Invalid limit parameter", "limit", limitString)
			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Limit must be greater than 0")
			return nil, false
		}

		return &parsedLimit, true
	} else {
		return nil, true
	}
}

func (ep *Endpoint) parseRowSelectionQueryParameters() *queries.RowSelection {
	limit, limitOk := ep.parseLimitQueryParameter()
	if !limitOk {
		return nil
	}

	offset, offsetOk := ep.parseOffsetQueryParameter()
	if !offsetOk {
		return nil
	}

	rowSelection := queries.RowSelection{
		Limit:  limit,
		Offset: offset,
	}
	return &rowSelection
}

func (ep *Endpoint) parseOrderQueryParameter() (queries.Order, bool) {
	if order, exists := ep.Context.GetQuery("order"); exists {
		if order != "antichronological" {
			ep.Logger.InvalidInput("Invalid order parameter", "order", order)
			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Order must be 'antichronological'")
			return 0, false
		}
		return queries.OrderAntiChronological, true
	}

	return queries.OrderChronological, true
}
