package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"

	"bctbackend/barcode"

	"github.com/go-pdf/fpdf"
)

type LabelData struct {
	BarcodeData      string
	Description      string
	Category         string
	ItemIdentifier   int
	PriceInCents     int
	SellerIdentifier int
	Charity          bool
	Donation         bool
}

type pdfBuilder struct {
	filename      string
	pdf           *fpdf.Fpdf
	imageCache    map[string]string
	layout        *ValidatedLayoutSettings
	gridWalker    *GridWalker
	labels        []LabelData
	barcodeWidth  int
	barcodeHeight int
}

func GeneratePdf(filename string, layout *ValidatedLayoutSettings, labels []LabelData) error {
	if !IsA4Size(layout) {
		return errors.New("only A4 paper size is supported")
	}

	orientation := "P"
	unit := "mm"
	paperSize := "A4"
	fontDirectory := ""
	pdf := fpdf.New(orientation, unit, paperSize, fontDirectory)

	builder := pdfBuilder{
		filename:      filename,
		imageCache:    make(map[string]string),
		pdf:           pdf,
		layout:        layout,
		gridWalker:    NewGridWalker(layout.columns, layout.rows),
		labels:        labels,
		barcodeWidth:  100,
		barcodeHeight: 20,
	}

	return builder.generateLabels()
}

func (builder *pdfBuilder) generateLabels() error {
	for _, label := range builder.labels {
		if builder.gridWalker.IsAtStart() {
			builder.pdf.AddPage()
		}

		rectangle := builder.layout.GetRectangle(builder.gridWalker.CurrentColumn, builder.gridWalker.CurrentRow)

		err := builder.generateLabel(&rectangle, &label)
		if err != nil {
			return err
		}

		builder.gridWalker.Next()
	}

	return builder.pdf.OutputFileAndClose(builder.filename)
}

func (builder *pdfBuilder) generateLabel(rectangle *Rectangle, labelData *LabelData) error {
	builder.pdf.ClipRect(rectangle.Left, rectangle.Top, rectangle.Width, rectangle.Height, false)
	defer builder.pdf.ClipEnd()

	builder.drawLabelBorder(rectangle)
	// builder.drawGrid(rectangle, 5)

	barcodeX := rectangle.Left + 2
	barcodeY := rectangle.Top + 2
	barcodeHeight, err := builder.drawBarcode(labelData.BarcodeData, barcodeX, barcodeY)

	if err != nil {
		return err
	}

	builder.pdf.AddUTF8Font("Arial", "", "Arial.ttf")
	builder.pdf.SetFont("Arial", "", 12)
	_, textHeightInMm := builder.pdf.GetFontSize()

	descriptionX := rectangle.Left + 2
	descriptionY := barcodeY + barcodeHeight + 2 + textHeightInMm
	builder.drawText(labelData.Description, descriptionX, descriptionY)

	categoryX := rectangle.Left + 2
	categoryY := descriptionY + textHeightInMm
	builder.drawText(labelData.Category, categoryX, categoryY)

	itemIdX := rectangle.Left + 2
	itemIdY := rectangle.Top + rectangle.Height - 2
	builder.drawText(fmt.Sprintf("%d", labelData.ItemIdentifier), itemIdX, itemIdY)

	priceAndSellerString := fmt.Sprintf("€%d.%02d → %d", labelData.PriceInCents/100, labelData.PriceInCents%100, labelData.SellerIdentifier)
	priceAndSellerWidth := builder.pdf.GetStringWidth(priceAndSellerString)
	priceAndSellerX := rectangle.Left + rectangle.Width - priceAndSellerWidth - 2
	priceAndSellerY := rectangle.Top + rectangle.Height - 2
	builder.drawText(priceAndSellerString, priceAndSellerX, priceAndSellerY)

	return nil
}

func (builder *pdfBuilder) generateBarcode(data string) (string, error) {
	if cached, ok := builder.imageCache[data]; ok {
		return cached, nil
	}

	// Generate barcode image in memory
	barcode, err := barcode.GenerateBarcode(data, builder.barcodeWidth, builder.barcodeHeight)
	if err != nil {
		return "", err
	}

	// Convert image to PNG format, still in memory
	var buffer bytes.Buffer
	err = png.Encode(&buffer, barcode)

	if err != nil {
		return "", err
	}

	// Generate image name
	imageIndex := len(builder.imageCache)
	imageName := fmt.Sprintf("barcode_%d", imageIndex)

	// Register image in PDF
	imageOptions := fpdf.ImageOptions{
		ImageType: "png",
		ReadDpi:   true,
	}
	builder.pdf.RegisterImageOptionsReader(imageName, imageOptions, &buffer)

	// Cache image
	builder.imageCache[data] = imageName

	// Return image name to be used when adding image to PDF
	return imageName, nil
}

func (builder *pdfBuilder) drawBarcode(data string, x float64, y float64) (float64, error) {
	imageName, err := builder.generateBarcode(data)
	if err != nil {
		return 0, err
	}

	imageOptions := fpdf.ImageOptions{
		ImageType: "png",
		ReadDpi:   true,
	}
	builder.pdf.ImageOptions(imageName, x, y, -1, -1, false, imageOptions, 0, "")

	imageInfo := builder.pdf.GetImageInfo(imageName)
	imageHeight := imageInfo.Height()

	return imageHeight, nil
}

func (builder *pdfBuilder) drawText(text string, x float64, y float64) {
	builder.pdf.Text(x, y, text)
}

func (builder *pdfBuilder) drawGrid(rectangle *Rectangle, cellSize float64) {
	r, g, b := builder.pdf.GetDrawColor()
	defer builder.pdf.SetDrawColor(r, g, b)

	builder.pdf.SetDrawColor(128, 128, 128)

	left := rectangle.Left
	top := rectangle.Top
	right := rectangle.Left + rectangle.Width
	bottom := rectangle.Top + rectangle.Height

	for dx := 0.0; dx < rectangle.Width; dx += cellSize {
		x := rectangle.Left + dx
		builder.pdf.Line(x, top, x, bottom)
	}

	for dy := 0.0; dy < rectangle.Height; dy += cellSize {
		y := rectangle.Top + dy
		builder.pdf.Line(left, y, right, y)
	}
}

func (builder *pdfBuilder) drawLabelBorder(rectangle *Rectangle) {
	builder.pdf.Rect(rectangle.Left, rectangle.Top, rectangle.Width, rectangle.Height, "D")
}
