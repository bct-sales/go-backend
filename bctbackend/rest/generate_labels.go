package rest

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	"bctbackend/pdf"
	"bctbackend/rest/failure_response"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
	ItemIds []models.Id `json:"itemIds"`
}

func GenerateLabels(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	if roleId != models.SellerRoleId {
		failure_response.WrongRole(context, "Only sellers can generate labels")
		return
	}

	var payload GenerateLabelsPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		failure_response.InvalidRequest(context, "Failed to parse payload:"+err.Error())
		return
	}

	labelData, err := collectLabelData(db, payload.ItemIds)
	if err != nil {
		failure_response.Unknown(context, "Failed to collect label data: "+err.Error())
		return
	}

	settings, err := pdf.NewLayoutSettings(
		pdf.WithPaperSize(
			payload.Layout.PaperWidth,
			payload.Layout.PaperHeight,
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
		pdf.WithFontSize(5),
	)
	if err != nil {
		// TODO Better error handling
		failure_response.InvalidRequest(context, "Failed to parse layout settings: "+err.Error())
		return
	}

	builder, err := pdf.GeneratePdf(settings, labelData)
	if err != nil {
		failure_response.InvalidRequest(context, "Failed to generate PDF: "+err.Error())
		return
	}

	buffer, err := builder.WriteToBuffer()
	if err != nil {
		failure_response.InvalidRequest(context, "Failed to write PDF to buffer: "+err.Error())
		return
	}

	if err := queries.UpdateFreezeStatusOfItems(db, payload.ItemIds, true); err != nil {
		failure_response.Unknown(context, "Failed to freeze items: "+err.Error())
		return
	}

	context.Header("Content-Disposition", "attachment; filename=labels.pdf")
	context.Header("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	context.Data(http.StatusOK, "application/pdf", buffer.Bytes())
}

func collectLabelData(db *sql.DB, itemIds []models.Id) ([]*pdf.LabelData, error) {
	itemTable, err := queries.GetItemsWithIds(db, itemIds)
	if err != nil {
		return nil, err
	}

	createLabelData := func(itemId models.Id) (*pdf.LabelData, error) {
		item, ok := itemTable[itemId]
		if !ok {
			return nil, fmt.Errorf("bug: item with id %d not found; should never occur: this error should have be caught earlier", itemId)
		}
		return createLabelDataFromItem(item)
	}

	labelData, err := algorithms.MapError(itemIds, createLabelData)
	if err != nil {
		return nil, err
	}

	return labelData, nil
}

func createLabelDataFromItem(item *models.Item) (*pdf.LabelData, error) {
	barcode := fmt.Sprintf("%dx", item.ItemId)

	category, err := defs.NameOfCategory(item.CategoryId)
	if err != nil {
		return nil, err
	}

	labelData := &pdf.LabelData{
		BarcodeData:      barcode,
		Description:      item.Description,
		Category:         category,
		ItemIdentifier:   int(item.ItemId),
		PriceInCents:     int(item.PriceInCents),
		SellerIdentifier: int(item.SellerId),
		Charity:          item.Charity,
		Donation:         item.Donation,
	}

	return labelData, nil
}
