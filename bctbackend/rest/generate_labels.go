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
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GenerateLabelsPayload struct {
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

	slog.Info("Generating labels", "itemIds", payload.ItemIds)

	createLabelData := func(itemId models.Id) (*pdf.LabelData, error) {
		return createLabelDataFromItemId(db, itemId)
	}
	labelData, err := algorithms.MapError(payload.ItemIds, createLabelData)
	if err != nil {
		failure_response.Unknown(context, "Failed to create label data: "+err.Error())
		return
	}

	settings, err := pdf.NewLayoutSettings(
		pdf.WithA4PaperSize(),
		pdf.WithGridSize(2, 8),
		pdf.WithUniformLabelMargin(2),
		pdf.WithUniformLabelPadding(2),
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

	context.Header("Content-Disposition", "attachment; filename=labels.pdf")
	context.Header("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	context.Data(http.StatusOK, "application/pdf", buffer.Bytes())
}

func createLabelDataFromItemId(db *sql.DB, itemId models.Id) (*pdf.LabelData, error) {
	item, err := queries.GetItemWithId(db, itemId)
	if err != nil {
		return nil, err
	}

	return createLabelDataFromItem(item)
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
