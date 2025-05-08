package barcode

import (
	"bctbackend/pdf"
)

func GeneratePdf() error {
	layout, err := pdf.NewLayoutSettings().SetA4PaperSize().SetPaperMargins(10.0).SetGridSize(2, 8).SetLabelMargin(2).SetLabelPadding(2).SetFontSize(5).Validate()

	if err != nil {
		return err
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
		return err
	}

	if err := result.WriteToFile("output.pdf"); err != nil {
		return err
	}

	return nil
}
