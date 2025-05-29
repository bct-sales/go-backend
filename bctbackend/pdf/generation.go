package pdf

import (
	"bytes"
	"fmt"
	"image/png"
	"os"

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
	builder, err := newPdfBuilder(layout, labels)
	if err != nil {
		return nil, err
	}

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
	fontDirectory := os.Getenv("BCT_FONT_DIR")

	return fpdf.New(orientation, unit, paperSize, fontDirectory)
}

func newPdfBuilder(layout *LayoutSettings, labels []*LabelData) (*PdfBuilder, error) {
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

	if err := builder.setFont(); err != nil {
		return nil, err
	}

	if err := builder.registerImages(); err != nil {
		return nil, err
	}

	return &builder, nil
}

func (builder *PdfBuilder) addPage() error {
	orientation := "P"
	pageSize := fpdf.SizeType{Wd: builder.layout.paperWidth, Ht: builder.layout.paperHeight}

	builder.pdf.AddPageFormat(orientation, pageSize)
	if err := builder.pdf.Error(); err != nil {
		return err
	}

	return nil
}

func (builder *PdfBuilder) registerImages() error {
	if err := builder.registerImage(donationImageName, DonationImageBuffer()); err != nil {
		return err
	}

	if err := builder.registerImage(charityImageName, CharityImageBuffer()); err != nil {
		return err
	}

	return nil
}

func (builder *PdfBuilder) drawLabels() error {
	for _, label := range builder.labels {
		if builder.gridWalker.IsAtStart() {
			if err := builder.addPage(); err != nil {
				return err
			}
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

	if err := builder.drawLabelBorder(labelRectangle); err != nil {
		return err
	}

	if builder.showGrid {
		if err := builder.drawGrid(labelRectangle, 5); err != nil {
			return err
		}
	}

	rectangle := labelRectangle.Shrink(builder.layout.labelPadding)

	barcodeX := rectangle.Left
	barcodeY := rectangle.Top
	barcodeImageName, err := builder.drawBarcode(labelData.BarcodeData, barcodeX, barcodeY)
	if err != nil {
		return err
	}

	_, barcodeHeight, err := builder.determineImageSize(barcodeImageName)
	if err != nil {
		return fmt.Errorf("failed to determine barcode size while drawing label: %w", err)
	}
	descriptionX := rectangle.Left
	descriptionY := barcodeY + barcodeHeight + builder.layout.fontSize
	if err := builder.drawText(labelData.Description, descriptionX, descriptionY); err != nil {
		return err
	}

	categoryX := rectangle.Left
	categoryY := descriptionY + builder.layout.fontSize
	if err := builder.drawText(labelData.Category, categoryX, categoryY); err != nil {
		return err
	}

	itemIdentifierString := fmt.Sprintf("%d", labelData.ItemIdentifier)
	if err := builder.drawTextInLowerLeftCorner(itemIdentifierString, rectangle); err != nil {
		return err
	}

	priceAndSellerString := formatPriceAndSeller(labelData.PriceInCents, labelData.SellerIdentifier)
	if err := builder.drawTextInLowerRightCorner(priceAndSellerString, rectangle); err != nil {
		return err
	}

	if labelData.Charity {
		if err := builder.drawCharityImage(rectangle); err != nil {
			return fmt.Errorf("failed to draw label: %w", err)
		}
	}

	if labelData.Donation {
		if err := builder.drawDonationImage(rectangle); err != nil {
			return fmt.Errorf("failed to draw label: %w", err)
		}
	}

	return nil
}

func (builder *PdfBuilder) drawCharityImage(rectangle *Rectangle) error {
	imageWidth, _, err := builder.determineImageSize(charityImageName)
	if err != nil {
		return fmt.Errorf("failed to draw charity image: %w", err)
	}

	x := rectangle.Right() - imageWidth
	y := rectangle.Top

	if err := builder.drawImage(charityImageName, x, y); err != nil {
		return fmt.Errorf("failed to draw charity image: %w", err)
	}

	return nil
}

func (builder *PdfBuilder) drawDonationImage(rectangle *Rectangle) error {
	charityImageWidth, _, err := builder.determineImageSize(charityImageName)
	if err != nil {
		return fmt.Errorf("failed to draw donation image: %w", err)
	}

	donationImageWidth, _, err := builder.determineImageSize(donationImageName)
	if err != nil {
		return fmt.Errorf("failed to draw donation image: %w", err)
	}

	x := rectangle.Right() - charityImageWidth - donationImageWidth - 2
	y := rectangle.Top

	if err := builder.drawImage(donationImageName, x, y); err != nil {
		return err
	}

	return nil
}

func formatPriceAndSeller(priceInCents int, sellerIdentifier int) string {
	euros := priceInCents / 100
	cents := priceInCents % 100

	return fmt.Sprintf("€%d.%02d → %d", euros, cents, sellerIdentifier)
}

func (builder *PdfBuilder) setFont() error {
	builder.pdf.AddUTF8Font("Arial", "", "Arial.ttf")
	if err := builder.pdf.Error(); err != nil {
		return err
	}

	fontSizeInPoints := builder.pdf.UnitToPointConvert(builder.layout.fontSize)
	if err := builder.pdf.Error(); err != nil {
		return err
	}

	builder.pdf.SetFont("Arial", "", fontSizeInPoints)
	if err := builder.pdf.Error(); err != nil {
		return err
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
		return err
	}

	return nil
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
		ImageType: "png",
		ReadDpi:   true,
	}

	builder.pdf.ImageOptions(imageName, x, y, -1, -1, false, imageOptions, 0, "")
	if err := builder.pdf.Error(); err != nil {
		return err
	}

	return nil
}

func (builder *PdfBuilder) determineImageSize(imageName string) (float64, float64, error) {
	imageInfo := builder.pdf.GetImageInfo(imageName)
	if imageInfo == nil {
		return 0, 0, fmt.Errorf("failed to get image info for %s", imageName)
	}

	imageWidth := imageInfo.Width()
	imageHeight := imageInfo.Height()

	return imageWidth, imageHeight, nil
}

func (builder *PdfBuilder) drawBarcode(data string, x float64, y float64) (string, error) {
	imageName, err := builder.generateBarcode(data)
	if err != nil {
		return "", err
	}

	if err := builder.drawImage(imageName, x, y); err != nil {
		return "", err
	}

	return imageName, nil
}

func (builder *PdfBuilder) drawText(text string, x float64, y float64) error {
	builder.pdf.Text(x, y, text)
	if err := builder.pdf.Error(); err != nil {
		return err
	}

	return nil
}

func (builder *PdfBuilder) drawTextInLowerLeftCorner(text string, rectangle *Rectangle) error {
	x := rectangle.Left
	y := rectangle.Bottom()

	builder.pdf.Text(x, y, text)
	if err := builder.pdf.Error(); err != nil {
		return err
	}

	return nil
}

func (builder *PdfBuilder) drawTextInLowerRightCorner(text string, rectangle *Rectangle) error {
	stringLength := builder.pdf.GetStringWidth(text)
	x := rectangle.Right() - stringLength
	y := rectangle.Bottom()

	builder.pdf.Text(x, y, text)
	if err := builder.pdf.Error(); err != nil {
		return err
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
		return err
	}

	return nil
}

func (builder *PdfBuilder) drawGrid(rectangle *Rectangle, cellSize float64) error {
	r, g, b := builder.pdf.GetDrawColor()
	defer builder.pdf.SetDrawColor(r, g, b)

	builder.pdf.SetDrawColor(128, 128, 128)
	if err := builder.pdf.Error(); err != nil {
		return err
	}

	left := rectangle.Left
	top := rectangle.Top
	right := rectangle.Right()
	bottom := rectangle.Bottom()

	for dx := 0.0; dx < rectangle.Width; dx += cellSize {
		x := rectangle.Left + dx
		builder.pdf.Line(x, top, x, bottom)
		if err := builder.pdf.Error(); err != nil {
			return err
		}
	}

	for dy := 0.0; dy < rectangle.Height; dy += cellSize {
		y := rectangle.Top + dy
		builder.pdf.Line(left, y, right, y)
		if err := builder.pdf.Error(); err != nil {
			return err
		}
	}

	return nil
}

func (builder *PdfBuilder) drawLabelBorder(rectangle *Rectangle) error {
	builder.pdf.Rect(rectangle.Left, rectangle.Top, rectangle.Width, rectangle.Height, "D")
	if err := builder.pdf.Error(); err != nil {
		return err
	}

	return nil
}
