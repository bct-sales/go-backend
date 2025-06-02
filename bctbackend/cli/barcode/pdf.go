package barcode

import (
	"bctbackend/pdf"
	"fmt"
)

func GeneratePdf() error {
	layout, err := pdf.NewLayoutSettings(
		pdf.WithA4PaperSize(),
		pdf.WithUniformPaperMargin(10.0),
		pdf.WithGridSize(2, 8),
		pdf.WithUniformLabelMargin(2),
		pdf.WithUniformLabelPadding(2),
		pdf.WithFontSize(5),
	)
	if err != nil {
		return fmt.Errorf("failed to create layout settings object: %w", err)
	}

	labels := []*pdf.LabelData{
		{
			BarcodeData:      "1x",
			Description:      "Test Product",
			Category:         "Test Category",
			ItemIdentifier:   1,
			PriceInCents:     1000,
			SellerIdentifier: 1,
			Charity:          true,
			Donation:         true,
		},
		{
			BarcodeData:      "2x",
			Description:      "Test Product2",
			Category:         "Test Category2",
			ItemIdentifier:   2,
			PriceInCents:     2000,
			SellerIdentifier: 2,
			Charity:          false,
			Donation:         false,
		},
	}

	result, err := pdf.GeneratePdf(layout, labels)
	if err != nil {
		return fmt.Errorf("failed to generate pdf: %w", err)
	}

	if err := result.WriteToFile("output.pdf"); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
