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
	"net/http"

	"log/slog"
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
	ItemIds []models.ID `json:"itemIds"`
}

type generateLabelsEndpoint struct {
	Endpoint
}

func GenerateLabels(arguments *HandlerFunctionArguments) {
	endpoint := &generateLabelsEndpoint{
		Endpoint: Endpoint{
			HandlerFunctionArguments: *arguments,
		},
	}

	endpoint.execute()
}

func (ep *generateLabelsEndpoint) execute() {
	if !ep.ensureUserIsSeller() {
		return
	}

	payload := ep.parsePayload()
	if payload == nil {
		return
	}

	if !ep.validatePayload(payload) {
		return
	}

	itemTable := ep.retrieveItemsFromDatabase(payload.ItemIds)
	if itemTable == nil {
		return
	}

	if !ep.checkItemOwnership(itemTable) {
		return
	}

	labelData := ep.collectLabelData(ep.Database, itemTable, payload.ItemIds)
	if labelData == nil {
		return
	}

	layoutSettings := ep.createLayoutSettings(payload)
	if layoutSettings == nil {
		return
	}

	buffer := ep.generatePdf(labelData, layoutSettings)
	if buffer == nil {
		return
	}

	// Do this last, to ensure items are only frozen if the PDF was generated successfully
	if !ep.freezeItems(ep.Database, payload.ItemIds) {
		return
	}

	ep.Context.Header("Content-Disposition", "attachment; filename=labels.pdf")
	ep.Context.DataFromReader(
		http.StatusOK,
		int64(buffer.Len()),
		"application/pdf",
		buffer,
		map[string]string{"Content-Disposition": "attachment; filename=labels.pdf"},
	)
}

func (ep *generateLabelsEndpoint) freezeItems(db Database, itemIds []models.ID) bool {
	transaction, err := db.StartTransaction()
	if err != nil {
		ep.Logger.InternalError("Failed to start transaction", "error", err)
		failure_response.Unknown(ep.Context, "Failed to start transaction while freezing items: "+err.Error())
		return false
	}
	defer transaction.RollbackIfNotCommitted()

	if err := queries.UpdateFreezeStatusOfItems(transaction, itemIds, true); err != nil {
		ep.Logger.InternalError("Failed to freeze items", "error", err)
		failure_response.Unknown(ep.Context, "Failed to freeze items: "+err.Error())
		return false
	}

	if err := transaction.Commit(); err != nil {
		ep.Logger.InternalError("Failed to commit transaction", "error", err)
		failure_response.Unknown(ep.Context, "Failed to commit transaction while freezing items: "+err.Error())
		return false
	}

	return true
}

func (ep *generateLabelsEndpoint) collectLabelData(db Database, itemTable map[models.ID]*models.Item, itemIds []models.ID) []*pdf.LabelData {
	createLabelData := func(itemId models.ID) (*pdf.LabelData, error) {
		item, ok := itemTable[itemId]
		if !ok {
			ep.Logger.Bug("Bug: did not find item with id %s", itemId.String())
			return nil, fmt.Errorf("bug: item with id %d not found; should never occur: this error should have be caught earlier", itemId)
		}

		categoryNameTable, err := queries.GetCategoryNameTable(db)
		if err != nil {
			ep.Logger.InternalError("Unable to build category table", itemId.String())
			return nil, err
		}

		return ep.createLabelDataFromItem(categoryNameTable, item)
	}

	labelData, err := algorithms.MapError(itemIds, createLabelData)
	if err != nil {
		ep.Logger.InternalError("Failed to collect label data", "error", err)
		failure_response.Unknown(ep.Context, "Failed to collect label data: "+err.Error())
		return nil
	}

	return labelData
}

func (ep *generateLabelsEndpoint) createLabelDataFromItem(categoryNameTable map[models.ID]string, item *models.Item) (*pdf.LabelData, error) {
	barcode := fmt.Sprintf("%dx", item.ItemID)

	category, ok := categoryNameTable[item.CategoryID]
	if !ok {
		ep.Logger.Bug("Unknown category %s", item.CategoryID)
		return nil, fmt.Errorf("unknown category id: %v", item.CategoryID)
	}

	labelData := &pdf.LabelData{
		BarcodeData:      barcode,
		Description:      item.Description,
		Category:         category,
		ItemIdentifier:   int(item.ItemID),
		PriceInCents:     int(item.PriceInCents),
		SellerIdentifier: int(item.SellerID),
		Charity:          item.Charity,
		Donation:         item.Donation,
	}

	return labelData, nil
}

func (ep *generateLabelsEndpoint) generatePdf(labelData []*pdf.LabelData, layoutSettings *pdf.LayoutSettings) *bytes.Buffer {
	pdfConfiguration := pdf.Configuration{
		FontDirectory: ep.Configuration.LabelGeneration.Font.Directory,
		FontFilename:  ep.Configuration.LabelGeneration.Font.Filename,
		FontFamily:    ep.Configuration.LabelGeneration.Font.Family,
		BarcodeWidth:  ep.Configuration.LabelGeneration.BarcodeWidth,
		BarcodeHeight: ep.Configuration.LabelGeneration.BarcodeHeight,
	}
	builder, err := pdf.GeneratePdf(&pdfConfiguration, layoutSettings, labelData)
	if err != nil {
		slog.Error("Failed to generate PDF", "error", err)
		failure_response.InvalidRequest(ep.Context, "Failed to generate PDF: "+err.Error())
		return nil
	}

	buffer, err := builder.WriteToBuffer()
	if err != nil {
		ep.Logger.InternalError("Failed to write PDF to buffer", "error", err)
		failure_response.InvalidRequest(ep.Context, "Failed to write PDF to buffer: "+err.Error())
		return nil
	}

	return buffer
}

func (ep *generateLabelsEndpoint) createLayoutSettings(payload *GenerateLabelsPayload) *pdf.LayoutSettings {
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
		ep.Logger.InvalidRequest("Invalid layout for label generation", "error", err)
		failure_response.InvalidLayout(ep.Context, "Invalid label layout: "+err.Error())
		return nil
	}

	return layoutSettings
}

func (ep *generateLabelsEndpoint) ensureUserIsSeller() bool {
	if !ep.RoleId.IsSeller() {
		ep.Logger.InvalidRequest("Blocked attempt at generating labels by a user with the wrong role")
		failure_response.WrongRole(ep.Context, "Only sellers can generate labels")
		return false
	}

	return true
}

func (ep *generateLabelsEndpoint) parsePayload() *GenerateLabelsPayload {
	var payload GenerateLabelsPayload

	if err := ep.Context.ShouldBindJSON(&payload); err != nil {
		ep.Logger.InvalidInput("Failed to parse payload for GenerateLabels endpoint", "error", err)
		failure_response.InvalidRequest(ep.Context, "Failed to parse payload:"+err.Error())
		return nil
	}

	return &payload
}

func (ep *generateLabelsEndpoint) validatePayload(payload *GenerateLabelsPayload) bool {
	if len(payload.ItemIds) == 0 {
		ep.Logger.InvalidRequest("Blocked attempt at generating labels for zero items", "userId", ep.UserId)
		failure_response.MissingItems(ep.Context, "No items provided")
		return false
	}

	return true
}

func (ep *generateLabelsEndpoint) retrieveItemsFromDatabase(itemIds []models.ID) map[models.ID]*models.Item {
	itemTable, err := queries.GetItemsWithIDs(ep.Database, itemIds)
	if err != nil {
		if errors.Is(err, dberr.ErrNoSuchItem) {
			ep.Logger.InvalidRequest("Blocked attempt at generating labels for non-existing items", "itemIds", itemIds, "userId", ep.UserId)
			failure_response.UnknownItem(ep.Context, err.Error())
			return nil
		}

		ep.Logger.InternalError("Failed to fetch items: %s", err.Error())
		failure_response.Unknown(ep.Context, "Failed to fetch items: "+err.Error())
		return nil
	}

	return itemTable
}

func (ep *generateLabelsEndpoint) checkItemOwnership(itemTable map[models.ID]*models.Item) bool {
	for _, item := range itemTable {
		if item.SellerID != ep.UserId {
			ep.Logger.InvalidRequest("Blocked attempt at generating labels for items not owned by the seller", "loggedInUserId", ep.UserId, "itemUserId", item.SellerID, "itemId", item.ItemID)
			failure_response.WrongSeller(ep.Context, "labels can only be generated by the owning seller")
			return false
		}
	}

	return true
}
