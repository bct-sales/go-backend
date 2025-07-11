package pdf

import (
	"bytes"
	"fmt"
	"image/png"
	"log/slog"

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
	showGrid      bool
	configuration *Configuration
}

type Configuration struct {
	FontDirectory string
	FontFilename  string
	FontFamily    string
	BarcodeWidth  int
	BarcodeHeight int
}

func GeneratePdf(configuration *Configuration, layout *LayoutSettings, labels []*LabelData) (*PdfBuilder, error) {
	builder, err := newPdfBuilder(configuration, layout, labels)
	if err != nil {
		return nil, &PdfError{Message: "failed to create pdf builder", Wrapped: err}
	}

	if err := builder.drawLabels(); err != nil {
		return nil, &PdfError{Message: "failed to draw labels", Wrapped: err}
	}

	return builder, nil
}

func (builder *PdfBuilder) WriteToFile(filename string) error {
	if err := builder.pdf.OutputFileAndClose(filename); err != nil {
		return &PdfError{Message: "failed to save PDF", Wrapped: err}
	}

	return nil
}

func (builder *PdfBuilder) WriteToBuffer() (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	if err := builder.pdf.Output(&buffer); err != nil {
		return nil, &PdfError{Message: "failed to save PDF to buffer", Wrapped: err}
	}
	return &buffer, nil
}

func newPdfGenerator(fontDirectory string) *fpdf.Fpdf {
	orientation := "P"
	unit := "mm"
	paperSize := "A4" // should not be used, new pages will have an explicitly specified size

	return fpdf.New(orientation, unit, paperSize, fontDirectory)
}

func newPdfBuilder(configuration *Configuration, layout *LayoutSettings, labels []*LabelData) (*PdfBuilder, error) {
	builder := PdfBuilder{
		imageCache:    make(map[string]string),
		pdf:           newPdfGenerator(configuration.FontDirectory),
		layout:        layout,
		gridWalker:    NewGridWalker(layout.columns, layout.rows),
		labels:        labels,
		showGrid:      false,
		configuration: configuration,
	}

	if err := builder.setFont(); err != nil {
		return nil, &PdfError{Message: "failed to set font", Wrapped: err}
	}

	if err := builder.registerImages(); err != nil {
		return nil, &PdfError{Message: "failed to register images", Wrapped: err}
	}

	return &builder, nil
}

func (builder *PdfBuilder) addPage() error {
	orientation := "P"
	pageSize := fpdf.SizeType{Wd: builder.layout.paperWidth, Ht: builder.layout.paperHeight}

	builder.pdf.AddPageFormat(orientation, pageSize)
	if err := builder.pdf.Error(); err != nil {
		return &PdfError{Message: "failed to add page format", Wrapped: err}
	}

	return nil
}

func (builder *PdfBuilder) registerImages() error {
	if err := builder.registerImage(donationImageName, DonationImageBuffer()); err != nil {
		return &PdfError{Message: "failed to register donation image", Wrapped: err}
	}

	if err := builder.registerImage(charityImageName, CharityImageBuffer()); err != nil {
		return &PdfError{Message: "failed to register charity image", Wrapped: err}
	}

	return nil
}

func (builder *PdfBuilder) drawLabels() error {
	for _, label := range builder.labels {
		if builder.gridWalker.IsAtStart() {
			if err := builder.addPage(); err != nil {
				return &PdfError{Message: "failed to add new page", Wrapped: err}
			}
		}

		rectangle := builder.layout.GetRectangle(builder.gridWalker.CurrentColumn, builder.gridWalker.CurrentRow)

		err := builder.drawLabel(rectangle, label)
		if err != nil {
			return &PdfError{Message: "failed to draw label", Wrapped: err}
		}

		builder.gridWalker.Next()
	}

	return nil
}

func (builder *PdfBuilder) drawLabel(labelRectangle *Rectangle, labelData *LabelData) error {
	builder.pdf.ClipRect(labelRectangle.Left, labelRectangle.Top, labelRectangle.Width, labelRectangle.Height, false)
	defer builder.pdf.ClipEnd()

	if err := builder.drawLabelBorder(labelRectangle); err != nil {
		return &PdfError{Message: "failed to draw label border", Wrapped: err}
	}

	if builder.showGrid {
		if err := builder.drawGrid(labelRectangle, 5); err != nil {
			return &PdfError{Message: "failed to draw grid", Wrapped: err}
		}
	}

	rectangle := labelRectangle.Shrink(builder.layout.labelPadding)

	barcodeX := rectangle.Left
	barcodeY := rectangle.Top
	barcodeImageName, err := builder.drawBarcode(labelData.BarcodeData, barcodeX, barcodeY)
	if err != nil {
		return &PdfError{Message: "failed to draw barcode", Wrapped: err}
	}

	_, barcodeHeight, err := builder.determineImageSize(barcodeImageName)
	if err != nil {
		return &PdfError{Message: "failed to determine barcode size", Wrapped: err}
	}
	descriptionX := rectangle.Left
	descriptionY := barcodeY + barcodeHeight + builder.layout.fontSize
	if err := builder.drawText(labelData.Description, descriptionX, descriptionY); err != nil {
		return &PdfError{Message: "failed to draw description", Wrapped: err}
	}

	categoryX := rectangle.Left
	categoryY := descriptionY + builder.layout.fontSize
	if err := builder.drawText(labelData.Category, categoryX, categoryY); err != nil {
		return &PdfError{Message: "failed to draw category", Wrapped: err}
	}

	itemIdentifierString := fmt.Sprintf("%d", labelData.ItemIdentifier)
	if err := builder.drawTextInLowerLeftCorner(itemIdentifierString, rectangle); err != nil {
		return &PdfError{Message: "failed to draw item identifier", Wrapped: err}
	}

	priceAndSellerString := formatPriceAndSeller(labelData.PriceInCents, labelData.SellerIdentifier)
	if err := builder.drawTextInLowerRightCorner(priceAndSellerString, rectangle); err != nil {
		return &PdfError{Message: "failed to draw price and seller", Wrapped: err}
	}

	if labelData.Charity {
		if err := builder.drawCharityImage(rectangle); err != nil {
			return &PdfError{Message: "failed to draw charity image", Wrapped: err}
		}
	}

	if labelData.Donation {
		if err := builder.drawDonationImage(rectangle); err != nil {
			return &PdfError{Message: "failed to draw donation image", Wrapped: err}
		}
	}

	return nil
}

func (builder *PdfBuilder) drawCharityImage(rectangle *Rectangle) error {
	imageWidth, _, err := builder.determineImageSize(charityImageName)
	if err != nil {
		return &PdfError{Message: "failed to determine charity image size", Wrapped: err}
	}

	x := rectangle.Right() - imageWidth
	y := rectangle.Top

	if err := builder.drawImage(charityImageName, x, y); err != nil {
		return &PdfError{Message: "failed to draw charity image", Wrapped: err}
	}

	return nil
}

func (builder *PdfBuilder) drawDonationImage(rectangle *Rectangle) error {
	charityImageWidth, _, err := builder.determineImageSize(charityImageName)
	if err != nil {
		return &PdfError{Message: "failed to determine charity image size", Wrapped: err}
	}

	donationImageWidth, _, err := builder.determineImageSize(donationImageName)
	if err != nil {
		return &PdfError{Message: "failed to determine donation image size", Wrapped: err}
	}

	x := rectangle.Right() - charityImageWidth - donationImageWidth - 2
	y := rectangle.Top

	if err := builder.drawImage(donationImageName, x, y); err != nil {
		return &PdfError{Message: "failed to draw donation image", Wrapped: err}
	}

	return nil
}

func formatPriceAndSeller(priceInCents int, sellerIdentifier int) string {
	euros := priceInCents / 100
	cents := priceInCents % 100

	return fmt.Sprintf("€%d.%02d → %d", euros, cents, sellerIdentifier)
}

func (builder *PdfBuilder) setFont() error {
	fontFamily := builder.configuration.FontFamily
	fontStyle := ""
	fontFilename := builder.configuration.FontFilename

	slog.Debug("Setting font", slog.String("family", fontFamily), slog.String("style", fontStyle), slog.String("filename", fontFilename))
	builder.pdf.AddUTF8Font(fontFamily, fontStyle, fontFilename)
	if err := builder.pdf.Error(); err != nil {
		slog.Error("Failed to set font", slog.String("error", err.Error()))
		return &PdfError{Message: "failed to add font", Wrapped: err}
	}

	slog.Debug("Converting font size to points", slog.Float64("fontSize", builder.layout.fontSize))
	fontSizeInPoints := builder.pdf.UnitToPointConvert(builder.layout.fontSize)
	if err := builder.pdf.Error(); err != nil {
		slog.Error("Failed to convert font size to points", slog.String("error", err.Error()))
		return &PdfError{Message: "failed to convert font size to points", Wrapped: err}
	}

	slog.Debug("Setting font", slog.String("family", fontFamily), slog.Float64("sizeInPoints", fontSizeInPoints))
	builder.pdf.SetFont(fontFamily, "", fontSizeInPoints)
	if err := builder.pdf.Error(); err != nil {
		slog.Debug("Failed to set font", slog.String("error", err.Error()))
		return &PdfError{Message: "failed to set font", Wrapped: err}
	}

	return nil
}

func (builder *PdfBuilder) registerImage(imageName string, imageBuffer *bytes.Buffer) error {
	//exhaustruct:ignore
	imageOptions := fpdf.ImageOptions{
		ImageType: "png",
		ReadDpi:   true,
	}

	builder.pdf.RegisterImageOptionsReader(imageName, imageOptions, imageBuffer)
	if err := builder.pdf.Error(); err != nil {
		return &PdfError{Message: fmt.Sprintf("failed to register image %s", imageName), Wrapped: err}
	}

	return nil
}

func (builder *PdfBuilder) generateBarcode(data string) (string, error) {
	if cached, ok := builder.imageCache[data]; ok {
		return cached, nil
	}

	// Generate barcode image in memory
	barcode, err := barcode.GenerateBarcode(data, builder.configuration.BarcodeWidth, builder.configuration.BarcodeHeight)
	if err != nil {
		return "", fmt.Errorf("failed to generate barcode: %w", err)
	}

	// Convert image to PNG format, still in memory
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, barcode); err != nil {
		return "", fmt.Errorf("failed to encode barcode as PNG: %w", err)
	}

	// Generate image name
	imageIndex := len(builder.imageCache)
	imageName := fmt.Sprintf("barcode_%d", imageIndex)

	// Register image
	if err := builder.registerImage(imageName, &buffer); err != nil {
		return "", err
	}

	// Cache image
	builder.imageCache[data] = imageName

	// Return image name to be used when adding image to PDF
	return imageName, nil
}

func (builder *PdfBuilder) drawImage(imageName string, x float64, y float64) error {
	imageOptions := fpdf.ImageOptions{
		ImageType:             "png",
		ReadDpi:               true,
		AllowNegativePosition: false,
	}

	builder.pdf.ImageOptions(imageName, x, y, -1, -1, false, imageOptions, 0, "")
	if err := builder.pdf.Error(); err != nil {
		return &PdfError{
			Message: fmt.Sprintf("failed to draw image %s", imageName),
			Wrapped: err,
		}
	}

	return nil
}

func (builder *PdfBuilder) determineImageSize(imageName string) (float64, float64, error) {
	imageInfo := builder.pdf.GetImageInfo(imageName)
	if imageInfo == nil {
		return 0, 0, &PdfError{
			Message: fmt.Sprintf("failed to get image information for %s", imageName),
			Wrapped: nil,
		}
	}

	imageWidth := imageInfo.Width()
	imageHeight := imageInfo.Height()

	return imageWidth, imageHeight, nil
}

func (builder *PdfBuilder) drawBarcode(data string, x float64, y float64) (string, error) {
	imageName, err := builder.generateBarcode(data)
	if err != nil {
		return "", &PdfError{
			Message: fmt.Sprintf("failed to generate barcode for data %s", data),
			Wrapped: err,
		}
	}

	if err := builder.drawImage(imageName, x, y); err != nil {
		return "", &PdfError{
			Message: "failed to draw barcode",
			Wrapped: err,
		}
	}

	return imageName, nil
}

func (builder *PdfBuilder) drawText(text string, x float64, y float64) error {
	builder.pdf.Text(x, y, text)
	if err := builder.pdf.Error(); err != nil {
		return &PdfError{Message: "failed to draw text", Wrapped: err}
	}

	return nil
}

func (builder *PdfBuilder) drawTextInLowerLeftCorner(text string, rectangle *Rectangle) error {
	x := rectangle.Left
	y := rectangle.Bottom()

	builder.pdf.Text(x, y, text)
	if err := builder.pdf.Error(); err != nil {
		return &PdfError{Message: "failed to draw text in lower left corner", Wrapped: err}
	}

	return nil
}

func (builder *PdfBuilder) drawTextInLowerRightCorner(text string, rectangle *Rectangle) error {
	stringLength := builder.pdf.GetStringWidth(text)
	x := rectangle.Right() - stringLength
	y := rectangle.Bottom()

	builder.pdf.Text(x, y, text)
	if err := builder.pdf.Error(); err != nil {
		return &PdfError{Message: "failed to draw text in lower right corner", Wrapped: err}
	}

	return nil
}

func (builder *PdfBuilder) drawTextInTopRightCorner(text string, rectangle *Rectangle) error {
	stringLength := builder.pdf.GetStringWidth(text)
	stringHeight := builder.layout.fontSize
	x := rectangle.Right() - stringLength
	y := rectangle.Top + stringHeight

	builder.pdf.Text(x, y, text)
	if err := builder.pdf.Error(); err != nil {
		return &PdfError{Message: "failed to draw text in top right corner", Wrapped: err}
	}

	return nil
}

func (builder *PdfBuilder) drawGrid(rectangle *Rectangle, cellSize float64) error {
	r, g, b := builder.pdf.GetDrawColor()
	defer builder.pdf.SetDrawColor(r, g, b)

	builder.pdf.SetDrawColor(128, 128, 128)
	if err := builder.pdf.Error(); err != nil {
		return &PdfError{Message: "failed to set draw color for grid", Wrapped: err}
	}

	left := rectangle.Left
	top := rectangle.Top
	right := rectangle.Right()
	bottom := rectangle.Bottom()

	for dx := 0.0; dx < rectangle.Width; dx += cellSize {
		x := rectangle.Left + dx
		builder.pdf.Line(x, top, x, bottom)
		if err := builder.pdf.Error(); err != nil {
			return &PdfError{Message: "failed to draw vertical grid line", Wrapped: err}
		}
	}

	for dy := 0.0; dy < rectangle.Height; dy += cellSize {
		y := rectangle.Top + dy
		builder.pdf.Line(left, y, right, y)
		if err := builder.pdf.Error(); err != nil {
			return &PdfError{Message: "failed to draw horizontal grid line", Wrapped: err}
		}
	}

	return nil
}

func (builder *PdfBuilder) drawLabelBorder(rectangle *Rectangle) error {
	builder.pdf.Rect(rectangle.Left, rectangle.Top, rectangle.Width, rectangle.Height, "D")
	if err := builder.pdf.Error(); err != nil {
		return &PdfError{Message: "failed to draw label border rectangle", Wrapped: err}
	}

	return nil
}
