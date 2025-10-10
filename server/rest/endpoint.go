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
	Logger        logger.Logger // This instance is unique to the request, meaning changes to the logger are localized
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
	offsetString := ep.Context.Query("offset")
	if offsetString != "" {
		parsedOffset, err := strconv.ParseUint(offsetString, 10, 64)

		if err != nil {
			ep.Logger.InvalidInput("Failed to parse offset", "error", err, "offset", offsetString)
			failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Failed to parse offset: "+err.Error())
			return nil, false
		}

		return &parsedOffset, true
	} else {
		return nil, true
	}
}

func (ep *Endpoint) parseLimitQueryParameter() (*uint64, bool) {
	limitString := ep.Context.Query("limit")
	if limitString != "" {
		parsedLimit, err := strconv.ParseUint(limitString, 10, 64)

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
	order := ep.Context.Query("order")
	switch order {
	case "", "chronological":
		return queries.OrderChronological, true
	case "antichronological":
		return queries.OrderAntiChronological, true
	default:
		ep.Logger.InvalidInput("Invalid order parameter", "order", order)
		failure_response.BadRequest(ep.Context, "invalid_uri_parameters", "Order must be either 'chronological' or 'antichronological'")
		return 0, false
	}
}

// parseBooleanQueryParameter looks for a query parameter with the given key.
// The value must be either true or false.
// If the value is missing, nil is returned.
func (ep *Endpoint) parseBooleanQueryParameter(key string) (*bool, bool) {
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
		errorMessage := fmt.Sprintf("Invalid query parameter value for %s", key)
		ep.Logger.InvalidInput(errorMessage)
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
		ep.Logger.InvalidInput("Unknown format requested", "format", requestedFormat)
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
