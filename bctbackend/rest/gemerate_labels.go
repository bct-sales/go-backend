package rest

import (
	"bctbackend/database/models"
	"bctbackend/pdf"
	"bctbackend/rest/failure_response"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GenerateLabelsPayload struct {
}

func GenerateLabels(context *gin.Context, db *sql.DB, userId models.Id, roleId models.Id) {
	// if roleId != models.SellerRoleId {
	// 	failure_response.WrongRole(context, "Only sellers can generate labels")
	// 	return
	// }

	// var payload GenerateLabelsPayload
	// if err := context.ShouldBindJSON(&payload); err != nil {
	// 	failure_response.InvalidRequest(context, "Failed to parse payload:"+err.Error())
	// 	return
	// }

	labelData := []pdf.LabelData{
		pdf.LabelData{
			BarcodeData:      "x",
			Description:      "Description",
			Category:         "Category",
			ItemIdentifier:   1,
			PriceInCents:     100,
			SellerIdentifier: 100,
			Charity:          true,
			Donation:         true,
		},
	}

	settings, err := pdf.NewLayoutSettings().SetA4PaperSize().SetPaperMargins(10.0).SetGridSize(2, 8).SetLabelMargin(2).SetLabelPadding(2).SetFontSize(5).Validate()
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
