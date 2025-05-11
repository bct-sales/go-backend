package pdf

import (
	"bytes"
	"fmt"
	"image/png"

	"bctbackend/barcode"

	"github.com/go-pdf/fpdf"
)

const (
	charityImageName  = "charity"
	donationImageName = "donation"
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

type PdfBuilder struct {
	pdf           *fpdf.Fpdf
	imageCache    map[string]string
	layout        *LayoutSettings
	gridWalker    *GridWalker
	labels        []*LabelData
	barcodeWidth  int
	barcodeHeight int
	showGrid      bool
}

func GeneratePdf(layout *LayoutSettings, labels []*LabelData) (*PdfBuilder, error) {
	builder := newPdfBuilder(layout, labels)

	if err := builder.drawLabels(); err != nil {
		return nil, err
	}

	return builder, nil
}

func (builder *PdfBuilder) WriteToFile(filename string) error {
	if err := builder.pdf.OutputFileAndClose(filename); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	return nil
}

func (builder *PdfBuilder) WriteToBuffer() (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	if err := builder.pdf.Output(&buffer); err != nil {
		return nil, fmt.Errorf("failed to save PDF to buffer: %w", err)
	}
	return &buffer, nil
}

func newPdfGenerator() *fpdf.Fpdf {
	orientation := "P"
	unit := "mm"
	paperSize := "A4" // should not be used, new pages will have an explicitly specified size
	fontDirectory := ""

	return fpdf.New(orientation, unit, paperSize, fontDirectory)
}

func newPdfBuilder(layout *LayoutSettings, labels []*LabelData) *PdfBuilder {
	builder := PdfBuilder{
		imageCache:    make(map[string]string),
		pdf:           newPdfGenerator(),
		layout:        layout,
		gridWalker:    NewGridWalker(layout.columns, layout.rows),
		labels:        labels,
		barcodeWidth:  100,
		barcodeHeight: 20,
		showGrid:      false,
	}

	builder.setFont()
	builder.registerImages()

	return &builder
}

func (builder *PdfBuilder) addPage() {
	orientation := "P"
	pageSize := fpdf.SizeType{Wd: builder.layout.paperWidth, Ht: builder.layout.paperHeight}
	builder.pdf.AddPageFormat(orientation, pageSize)
}

func (builder *PdfBuilder) registerImages() {
	builder.registerImage(donationImageName, DonationImageBuffer())
	builder.registerImage(charityImageName, CharityImageBuffer())
}

func (builder *PdfBuilder) drawLabels() error {
	for _, label := range builder.labels {
		if builder.gridWalker.IsAtStart() {
			builder.addPage()
		}

		rectangle := builder.layout.GetRectangle(builder.gridWalker.CurrentColumn, builder.gridWalker.CurrentRow)

		err := builder.drawLabel(rectangle, label)
		if err != nil {
			return err
		}

		builder.gridWalker.Next()
	}

	return nil
}

func (builder *PdfBuilder) drawLabel(labelRectangle *Rectangle, labelData *LabelData) error {
	if labelRectangle == nil {
		return fmt.Errorf("label rectangle is nil")
	}

	if labelData == nil {
		return fmt.Errorf("label data is nil")
	}

	builder.pdf.ClipRect(labelRectangle.Left, labelRectangle.Top, labelRectangle.Width, labelRectangle.Height, false)
	defer builder.pdf.ClipEnd()

	builder.drawLabelBorder(labelRectangle)

	if builder.showGrid {
		builder.drawGrid(labelRectangle, 5)
	}

	rectangle := labelRectangle.Shrink(builder.layout.labelPadding)

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

	if labelData.Charity {
		builder.drawCharityImage(rectangle)
	}

	if labelData.Donation {
		builder.drawDonationImage(rectangle)
	}

	return nil
}

func (builder *PdfBuilder) drawCharityImage(rectangle *Rectangle) {
	imageWidth, _ := builder.determineImageSize(charityImageName)
	x := rectangle.Right() - imageWidth
	y := rectangle.Top

	builder.drawImage(charityImageName, x, y)
}

func (builder *PdfBuilder) drawDonationImage(rectangle *Rectangle) {
	charityImageWidth, _ := builder.determineImageSize(charityImageName)
	donationImageWidth, _ := builder.determineImageSize(donationImageName)
	x := rectangle.Right() - charityImageWidth - donationImageWidth - 2
	y := rectangle.Top

	builder.drawImage(donationImageName, x, y)
}

func formatPriceAndSeller(priceInCents int, sellerIdentifier int) string {
	euros := priceInCents / 100
	cents := priceInCents % 100

	return fmt.Sprintf("€%d.%02d → %d", euros, cents, sellerIdentifier)
}

func (builder *PdfBuilder) setFont() {
	builder.pdf.AddUTF8Font("Arial", "", "Arial.otf")
	fontSizeInPoints := builder.pdf.UnitToPointConvert(builder.layout.fontSize)
	builder.pdf.SetFont("Arial", "", fontSizeInPoints)
}

func (builder *PdfBuilder) registerImage(imageName string, imageBuffer *bytes.Buffer) {
	imageOptions := fpdf.ImageOptions{
		ImageType: "png",
		ReadDpi:   true,
	}

	builder.pdf.RegisterImageOptionsReader(imageName, imageOptions, imageBuffer)
}

func (builder *PdfBuilder) generateBarcode(data string) (string, error) {
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

func (builder *PdfBuilder) drawImage(imageName string, x float64, y float64) {
	imageOptions := fpdf.ImageOptions{
		ImageType: "png",
		ReadDpi:   true,
	}

	builder.pdf.ImageOptions(imageName, x, y, -1, -1, false, imageOptions, 0, "")
}

func (builder *PdfBuilder) determineImageSize(imageName string) (float64, float64) {
	imageInfo := builder.pdf.GetImageInfo(imageName)
	imageWidth := imageInfo.Width()
	imageHeight := imageInfo.Height()

	return imageWidth, imageHeight
}

func (builder *PdfBuilder) drawBarcode(data string, x float64, y float64) (string, error) {
	imageName, err := builder.generateBarcode(data)
	if err != nil {
		return "", err
	}

	builder.drawImage(imageName, x, y)

	return imageName, nil
}

func (builder *PdfBuilder) drawText(text string, x float64, y float64) {
	builder.pdf.Text(x, y, text)
}

func (builder *PdfBuilder) drawTextInLowerLeftCorner(text string, rectangle *Rectangle) {
	x := rectangle.Left
	y := rectangle.Bottom()
	builder.pdf.Text(x, y, text)
}

func (builder *PdfBuilder) drawTextInLowerRightCorner(text string, rectangle *Rectangle) {
	stringLength := builder.pdf.GetStringWidth(text)
	x := rectangle.Right() - stringLength
	y := rectangle.Bottom()
	builder.pdf.Text(x, y, text)
}

func (builder *PdfBuilder) drawTextInTopRightCorner(text string, rectangle *Rectangle) {
	stringLength := builder.pdf.GetStringWidth(text)
	stringHeight := builder.layout.fontSize
	x := rectangle.Right() - stringLength
	y := rectangle.Top + stringHeight
	builder.pdf.Text(x, y, text)
}

func (builder *PdfBuilder) drawGrid(rectangle *Rectangle, cellSize float64) {
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

func (builder *PdfBuilder) drawLabelBorder(rectangle *Rectangle) {
	builder.pdf.Rect(rectangle.Left, rectangle.Top, rectangle.Width, rectangle.Height, "D")
}
