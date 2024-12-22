package barcode

import (
	"bctbackend/pdf"
)

func GeneratePdf() error {
	layout := pdf.NewLayoutSettings().SetA4PaperSize().SetPaperMargins(10.0).SetGridSize(2, 8).Validate()

	return pdf.GeneratePdf("output.pdf", layout)
}
