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
	showGrid      bool
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
		showGrid:      false,
	}

	return builder.generateLabels()
}

func (builder *pdfBuilder) generateLabels() error {
	for _, label := range builder.labels {
		if builder.gridWalker.IsAtStart() {
			builder.pdf.AddPage()
		}

		rectangle := builder.layout.GetRectangle(builder.gridWalker.CurrentColumn, builder.gridWalker.CurrentRow)

		err := builder.generateLabel(rectangle, &label)
		if err != nil {
			return err
		}

		builder.gridWalker.Next()
	}

	return builder.pdf.OutputFileAndClose(builder.filename)
}

func (builder *pdfBuilder) generateLabel(labelRectangle *Rectangle, labelData *LabelData) error {
	builder.pdf.ClipRect(labelRectangle.Left, labelRectangle.Top, labelRectangle.Width, labelRectangle.Height, false)
	defer builder.pdf.ClipEnd()

	builder.drawLabelBorder(labelRectangle)

	if builder.showGrid {
		builder.drawGrid(labelRectangle, 5)
	}

	rectangle := labelRectangle.ShrinkUniformly(builder.layout.labelPadding)

	builder.setFont()

	barcodeX := rectangle.Left
	barcodeY := rectangle.Top
	barcodeImageName, err := builder.drawBarcode(labelData.BarcodeData, barcodeX, barcodeY)

	if err != nil {
		return err
	}

	_, barcodeHeight := builder.determineImageSize(barcodeImageName)
	descriptionX := rectangle.Left
	descriptionY := barcodeY + barcodeHeight + builder.layout.fontSize
	builder.drawText(labelData.Description, descriptionX, descriptionY)

	categoryX := rectangle.Left
	categoryY := descriptionY + builder.layout.fontSize
	builder.drawText(labelData.Category, categoryX, categoryY)

	itemIdentifierString := fmt.Sprintf("%d", labelData.ItemIdentifier)
	builder.drawTextInLowerLeftCorner(itemIdentifierString, rectangle)

	priceAndSellerString := formatPriceAndSeller(labelData.PriceInCents, labelData.SellerIdentifier)
	builder.drawTextInLowerRightCorner(priceAndSellerString, rectangle)

	return nil
}

func formatPriceAndSeller(priceInCents int, sellerIdentifier int) string {
	euros := priceInCents / 100
	cents := priceInCents % 100

	return fmt.Sprintf("€%d.%02d → %d", euros, cents, sellerIdentifier)
}

func (builder *pdfBuilder) setFont() {
	builder.pdf.AddUTF8Font("Arial", "", "Arial.otf")
	fontSizeInPoints := builder.pdf.UnitToPointConvert(builder.layout.fontSize)
	builder.pdf.SetFont("Arial", "", fontSizeInPoints)
}

func (builder *pdfBuilder) registerImage(imageName string, imageBuffer *bytes.Buffer) {
	imageOptions := fpdf.ImageOptions{
		ImageType: "png",
		ReadDpi:   true,
	}

	builder.pdf.RegisterImageOptionsReader(imageName, imageOptions, imageBuffer)
}

func (builder *pdfBuilder) generateBarcode(data string) (string, error) {
	if cached, ok := builder.imageCache[data]; ok {
		return cached, nil
	}

	// Generate barcode image in memory
	barcode, err := barcode.GenerateBarcode(data, builder.barcodeWidth, builder.barcodeHeight)
	if err != nil {
		return "", fmt.Errorf("failed to generate barcode: %w", err)
	}

	// Convert image to PNG format, still in memory
	var buffer bytes.Buffer
	err = png.Encode(&buffer, barcode)

	if err != nil {
		return "", fmt.Errorf("failed to encode barcode as PNG: %w", err)
	}

	// Generate image name
	imageIndex := len(builder.imageCache)
	imageName := fmt.Sprintf("barcode_%d", imageIndex)

	// Register image
	builder.registerImage(imageName, &buffer)

	// Cache image
	builder.imageCache[data] = imageName

	// Return image name to be used when adding image to PDF
	return imageName, nil
}

func (builder *pdfBuilder) drawImage(imageName string, x float64, y float64) {
	imageOptions := fpdf.ImageOptions{
		ImageType: "png",
		ReadDpi:   true,
	}

	builder.pdf.ImageOptions(imageName, x, y, -1, -1, false, imageOptions, 0, "")
}

func (builder *pdfBuilder) determineImageSize(imageName string) (float64, float64) {
	imageInfo := builder.pdf.GetImageInfo(imageName)
	imageWidth := imageInfo.Width()
	imageHeight := imageInfo.Height()

	return imageWidth, imageHeight
}

func (builder *pdfBuilder) drawBarcode(data string, x float64, y float64) (string, error) {
	imageName, err := builder.generateBarcode(data)
	if err != nil {
		return "", err
	}

	builder.drawImage(imageName, x, y)

	return imageName, nil
}

func (builder *pdfBuilder) drawText(text string, x float64, y float64) {
	builder.pdf.Text(x, y, text)
}

func (builder *pdfBuilder) drawTextInLowerLeftCorner(text string, rectangle *Rectangle) {
	x := rectangle.Left
	y := rectangle.Bottom()
	builder.pdf.Text(x, y, text)
}

func (builder *pdfBuilder) drawTextInLowerRightCorner(text string, rectangle *Rectangle) {
	stringLength := builder.pdf.GetStringWidth(text)
	x := rectangle.Right() - stringLength
	y := rectangle.Bottom()
	builder.pdf.Text(x, y, text)
}

func (builder *pdfBuilder) drawTextInTopRightCorner(text string, rectangle *Rectangle) {
	stringLength := builder.pdf.GetStringWidth(text)
	stringHeight := builder.layout.fontSize
	x := rectangle.Right() - stringLength
	y := rectangle.Top + stringHeight
	builder.pdf.Text(x, y, text)
}

func (builder *pdfBuilder) drawGrid(rectangle *Rectangle, cellSize float64) {
	r, g, b := builder.pdf.GetDrawColor()
	defer builder.pdf.SetDrawColor(r, g, b)

	builder.pdf.SetDrawColor(128, 128, 128)

	left := rectangle.Left
	top := rectangle.Top
	right := rectangle.Right()
	bottom := rectangle.Bottom()

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
