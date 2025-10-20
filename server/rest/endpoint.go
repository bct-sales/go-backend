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
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HandlerFunctionArguments struct {
	Clock         clock.Clock
	Context       *gin.Context
	Configuration *configuration.Configuration
	Database      Database
	UserID        models.ID
	RoleID        models.RoleID
	Logger        logger.RestLogger // This instance is unique to the request, meaning changes to the logger are localized
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
	return queries.NewTransactionalDatabaseQuerier(w.Context, w.Database)
}

type Endpoint struct {
	HandlerFunctionArguments
}

// parseOffsetQueryParameter looks for a query parameter named offset.
// If it is missing, nil, true is returned.
// If it is present and its value is a valid integer, the parsed integer and true are returned.
// If it is present and its value is invalid, nil, false is returned.
func (ep *Endpoint) parseOffsetQueryParameter() (*uint64, bool) {
	logger := ep.Logger
	offsetString := ep.Context.Query("offset")

	if offsetString != "" {
		parsedOffset, err := strconv.ParseUint(offsetString, 10, 64)

		if err != nil {
			logger.AddInformation("offset", offsetString)
			logger.AddInformation("error", err)
			logger.InvalidInput("Failed to parse offset")

			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Failed to parse offset: "+err.Error())

			return nil, false
		}

		logger.AddInformation("offset", parsedOffset)

		return &parsedOffset, true
	} else {
		return nil, true
	}
}

func (ep *Endpoint) parseLimitQueryParameter() (*uint64, bool) {
	logger := ep.Logger
	limitString := ep.Context.Query("limit")

	if limitString != "" {
		parsedLimit, err := strconv.ParseUint(limitString, 10, 64)

		if err != nil {
			logger.AddInformation("limit", limitString)
			logger.AddInformation("error", err)
			logger.InvalidInput("Failed to parse limit")

			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Failed to parse limit: "+err.Error())

			return nil, false
		}

		if parsedLimit < 1 {
			logger.AddInformation("limit", limitString)
			logger.InvalidRequest("Invalid limit parameter")

			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Limit must be greater than 0")

			return nil, false
		}

		logger.AddInformation("limit", parsedLimit)

		return &parsedLimit, true
	} else {
		return nil, true
	}
}

func (ep *Endpoint) parseRowRangeQueryParameters() *queries.RowRange {
	limit, limitOk := ep.parseLimitQueryParameter()
	if !limitOk {
		return nil
	}

	offset, offsetOk := ep.parseOffsetQueryParameter()
	if !offsetOk {
		return nil
	}

	rowSelection := queries.RowRange{
		Limit:  limit,
		Offset: offset,
	}

	return &rowSelection
}

func (ep *Endpoint) parseOrderQueryParameter() (queries.Order, bool) {
	logger := ep.Logger
	order := ep.Context.Query("order")

	switch order {
	case "", "chronological":
		return queries.OrderChronological, true
	case "antichronological":
		return queries.OrderAntiChronological, true
	default:
		logger.AddInformation("order", order)
		logger.InvalidInput("Invalid order parameter")

		failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Order must be either 'chronological' or 'antichronological'")

		return 0, false
	}
}

// parseBooleanQueryParameter looks for a query parameter with the given key.
// The value must be either true or false.
// If the value is missing, nil is returned.
func (ep *Endpoint) parseBooleanQueryParameter(key string) (*bool, bool) {
	logger := ep.Logger
	value := ep.Context.Query(key)

	switch value {
	case "true":
		var value = true
		return &value, true

	case "false":
		var value = false
		return &value, true

	case "":
		return nil, true

	default:
		logger.AddInformation("query parameter key", key)
		logger.AddInformation("query parameter value", value)
		logger.InvalidInput("Invalid query parameter value")

		failure_response.BadRequest(ep.Context, "invalid_uri_parameters", fmt.Sprintf("%s must be either true or false", key))

		return nil, false
	}
}

func (ep *Endpoint) parseHiddenQueryParameter() (*bool, bool) {
	return ep.parseBooleanQueryParameter("hidden")
}

type formatHandler interface {
	handleDefaultFormat()
	handleCSVFormat()
	handleJSONFormat()
}

func (ep *Endpoint) parseFormatQueryParameter(formatHandler formatHandler) {
	logger := ep.Logger
	requestedFormat := ep.Context.Query("format")

	switch requestedFormat {
	case "":
		formatHandler.handleDefaultFormat()
		return

	case "json":
		formatHandler.handleJSONFormat()
		return

	case "csv":
		formatHandler.handleCSVFormat()
		return

	default:
		logger.AddInformation("format", requestedFormat)
		logger.InvalidInput("Unknown format requested")

		failure_response.Unknown(ep.Context, "Unknown format: "+requestedFormat)

		return
	}
}

type formatHandlerAdapter struct {
	handleDefaultFormatFunc func()
	handleCSVFormatFunc     func()
	handleJSONFormatFunc    func()
}

func (fha *formatHandlerAdapter) handleDefaultFormat() {
	fha.handleDefaultFormatFunc()
}

func (fha *formatHandlerAdapter) handleCSVFormat() {
	fha.handleCSVFormatFunc()
}

func (fha *formatHandlerAdapter) handleJSONFormat() {
	fha.handleJSONFormatFunc()
}
