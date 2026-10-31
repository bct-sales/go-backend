package rest

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/pdf"
	"bctbackend/server/failure_response"
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Insets struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

type Layout struct {
	PaperWidth   float64 `json:"paperWidth"`
	PaperHeight  float64 `json:"paperHeight"`
	PaperMargins Insets  `json:"paperMargins"`
	Columns      int     `json:"columns"`
	Rows         int     `json:"rows"`
	LabelMargins Insets  `json:"labelMargins"`
	LabelPadding Insets  `json:"labelPadding"`
	FontSize     float64 `json:"fontSize"`
}

type GenerateLabelsPayload struct {
	Layout  Layout      `json:"layout"`
	ItemIDs []models.ID `json:"itemIds"`
}

type generateLabelsEndpoint struct {
	Endpoint
}

func GenerateLabels(context *gin.Context, arguments *HandlerFunctionArguments) {
	endpoint := &generateLabelsEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute(context)
}

func (ep *generateLabelsEndpoint) execute(context *gin.Context) {
	if !ep.ensureUserIsSeller(context) {
		return
	}

	payload := ep.parsePayload(context)
	if payload == nil {
		return
	}

	if !ep.validatePayload(context, payload) {
		return
	}

	itemTable := ep.retrieveItemsFromDatabase(context, payload.ItemIDs)
	if itemTable == nil {
		return
	}

	if !ep.checkItemOwnership(context, itemTable) {
		return
	}

	labelData := ep.collectLabelData(context, ep.Database, itemTable, payload.ItemIDs)
	if labelData == nil {
		return
	}

	layoutSettings := ep.createLayoutSettings(context, payload)
	if layoutSettings == nil {
		return
	}

	buffer := ep.generatePdf(context, labelData, layoutSettings)
	if buffer == nil {
		return
	}

	// Do this last, to ensure items are only frozen if the PDF was generated successfully
	if !ep.freezeItems(context, ep.Database, payload.ItemIDs) {
		return
	}

	context.Header("Content-Disposition", "attachment; filename=labels.pdf")
	context.DataFromReader(
		http.StatusOK,
		int64(buffer.Len()),
		"application/pdf",
		buffer,
		map[string]string{"Content-Disposition": "attachment; filename=labels.pdf"},
	)
}

func (ep *generateLabelsEndpoint) freezeItems(context *gin.Context, db Database, itemIDs []models.ID) bool {
	logger := ep.Logger
	ctx := context.Request.Context()
	transaction, err := db.StartTransaction(ctx)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to start transaction")

		failure_response.Unknown(context, "Failed to start transaction while freezing items: "+err.Error())

		return false
	}
	defer transaction.RollbackIfNotCommitted()

	if err := queries.UpdateFreezeStatusOfItems(ctx, transaction, itemIDs, true); err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to freeze items")

		failure_response.Unknown(context, "Failed to freeze items: "+err.Error())

		return false
	}

	if err := transaction.Commit(); err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to commit transaction")

		failure_response.Unknown(context, "Failed to commit transaction while freezing items: "+err.Error())

		return false
	}

	return true
}

func (ep *generateLabelsEndpoint) collectLabelData(context *gin.Context, db Database, itemTable map[models.ID]*models.Item, itemIDs []models.ID) []*pdf.LabelData {
	logger := ep.Logger
	ctx := context.Request.Context()

	createLabelData := func(itemID models.ID) (*pdf.LabelData, error) {
		item, ok := itemTable[itemID]
		if !ok {
			logger.AddInformation("itemID", itemID)
			logger.Bug("Bug: did not find item")

			return nil, fmt.Errorf("bug: item with id %d not found; should never occur: this error should have be caught earlier", itemID)
		}

		categoryNameTable, err := queries.GetCategoryNameTable(ctx, db)
		if err != nil {
			logger.AddInformation("error", err)
			logger.InternalError("Unable to build category table")

			return nil, err
		}

		labelData, err := ep.createLabelDataFromItem(categoryNameTable, item)
		if err != nil {
			logger.AddInformation("itemID", itemID)
			logger.AddInformation("error", err)
			logger.InternalError("Unable to create label data for item")

			return nil, err
		}

		return labelData, nil
	}

	labelData, err := algorithms.MapError(itemIDs, createLabelData)
	if err != nil {
		return nil
	}

	return labelData
}

func (ep *generateLabelsEndpoint) createLabelDataFromItem(categoryNameTable map[models.ID]string, item *models.Item) (*pdf.LabelData, error) {
	logger := ep.Logger
	barcode := fmt.Sprintf("%dx", item.ItemID)

	category, ok := categoryNameTable[item.CategoryID]
	if !ok {
		logger.AddInformation("categoryID", item.CategoryID)
		logger.Bug("Unknown category")

		return nil, fmt.Errorf("unknown category id: %v", item.CategoryID)
	}

	labelData := &pdf.LabelData{
		BarcodeData:  barcode,
		Description:  item.Description,
		Category:     category,
		ItemID:       int(item.ItemID),
		PriceInCents: int(item.PriceInCents),
		SellerID:     int(item.SellerID),
		Charity:      item.Charity,
		Donation:     item.Donation,
		LargeItem:    item.Large,
	}

	return labelData, nil
}

func (ep *generateLabelsEndpoint) generatePdf(context *gin.Context, labelData []*pdf.LabelData, layoutSettings *pdf.LayoutSettings) *bytes.Buffer {
	logger := ep.Logger

	pdfConfiguration := pdf.Configuration{
		BarcodeWidth:  ep.Configuration.LabelGeneration.BarcodeWidth,
		BarcodeHeight: ep.Configuration.LabelGeneration.BarcodeHeight,
	}
	builder, err := pdf.GeneratePdf(&pdfConfiguration, layoutSettings, labelData)

	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to generate PDF")

		failure_response.InvalidRequest(context, "Failed to generate PDF: "+err.Error())

		return nil
	}

	buffer, err := builder.WriteToBuffer()
	if err != nil {
		logger.AddInformation("error", err)
		logger.InternalError("Failed to write PDF to buffer")

		failure_response.InvalidRequest(context, "Failed to write PDF to buffer: "+err.Error())

		return nil
	}

	return buffer
}

func (ep *generateLabelsEndpoint) createLayoutSettings(context *gin.Context, payload *GenerateLabelsPayload) *pdf.LayoutSettings {
	logger := ep.Logger

	layoutSettings, err := pdf.NewLayoutSettings(
		pdf.WithPaperSize(
			payload.Layout.PaperWidth,
			payload.Layout.PaperHeight,
		),
		pdf.WithPaperMargins(
			payload.Layout.PaperMargins.Top,
			payload.Layout.PaperMargins.Right,
			payload.Layout.PaperMargins.Bottom,
			payload.Layout.PaperMargins.Left,
		),
		pdf.WithGridSize(payload.Layout.Columns, payload.Layout.Rows),
		pdf.WithLabelMargins(
			payload.Layout.LabelMargins.Top,
			payload.Layout.LabelMargins.Right,
			payload.Layout.LabelMargins.Bottom,
			payload.Layout.LabelMargins.Left,
		),
		pdf.WithLabelPadding(
			payload.Layout.LabelPadding.Top,
			payload.Layout.LabelPadding.Right,
			payload.Layout.LabelPadding.Bottom,
			payload.Layout.LabelPadding.Left,
		),
		pdf.WithFontSize(payload.Layout.FontSize),
	)
	if err != nil {
		logger.AddInformation("error", err)
		logger.InvalidRequest("Invalid layout for label generation")

		failure_response.InvalidLayout(context, "Invalid label layout: "+err.Error())

		return nil
	}

	return layoutSettings
}

func (ep *generateLabelsEndpoint) ensureUserIsSeller(context *gin.Context) bool {
	logger := ep.Logger

	if !ep.RoleID.IsSeller() {
		logger.AddInformation("roleID", ep.RoleID)
		logger.InvalidRequest("Blocked attempt at generating labels by a user with the wrong role")

		failure_response.WrongRole(context, "Only sellers can generate labels")

		return false
	}

	return true
}

func (ep *generateLabelsEndpoint) parsePayload(context *gin.Context) *GenerateLabelsPayload {
	logger := ep.Logger
	var payload GenerateLabelsPayload

	if err := context.ShouldBindJSON(&payload); err != nil {
		logger.AddInformation("error", err)
		logger.InvalidInput("Failed to parse payload for GenerateLabels endpoint")

		failure_response.InvalidRequest(context, "Failed to parse payload:"+err.Error())

		return nil
	}

	ep.Logger.AddInformation("payload", payload)

	return &payload
}

func (ep *generateLabelsEndpoint) validatePayload(context *gin.Context, payload *GenerateLabelsPayload) bool {
	logger := ep.Logger

	if len(payload.ItemIDs) == 0 {
		logger.AddInformation("userID", ep.UserID)
		logger.InvalidRequest("Blocked attempt at generating labels for zero items")

		failure_response.MissingItems(context, "No items provided")

		return false
	}

	return true
}

func (ep *generateLabelsEndpoint) retrieveItemsFromDatabase(context *gin.Context, itemIDs []models.ID) map[models.ID]*models.Item {
	logger := ep.Logger
	ctx := context.Request.Context()
	itemTable, err := queries.GetItemsWithIDs(ctx, ep.Database, itemIDs)

	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchItem) {
			logger.AddInformation("itemIDs", itemIDs)
			logger.AddInformation("userID", ep.UserID)
			logger.InvalidRequest("Blocked attempt at generating labels for non-existing items")

			failure_response.UnknownItem(context, err.Error())

			return nil
		}

		logger.AddInformation("error", err)
		logger.InternalError("Failed to fetch items")

		failure_response.Unknown(context, "Failed to fetch items: "+err.Error())

		return nil
	}

	return itemTable
}

func (ep *generateLabelsEndpoint) checkItemOwnership(context *gin.Context, itemTable map[models.ID]*models.Item) bool {
	logger := ep.Logger

	for _, item := range itemTable {
		if item.SellerID != ep.UserID {
			logger.AddInformation("logged in user", ep.UserID)
			logger.AddInformation("item owner", item.SellerID)
			logger.AddInformation("itemID", item.ItemID)
			logger.InvalidRequest("Blocked attempt at generating labels for items not owned by the seller")

			failure_response.WrongSeller(context, "labels can only be generated by the owning seller")

			return false
		}
	}

	return true
}
