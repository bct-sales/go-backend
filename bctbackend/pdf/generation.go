package pdf

import (
	"errors"

	"github.com/go-pdf/fpdf"
)

func GeneratePdf(filename string, layout *ValidatedLayoutSettings) error {
	if !isA4Size(layout) {
		return errors.New("only A4 paper size is supported")
	}

	orientation := "P"
	unit := "mm"
	paperSize := "A4"
	fontDirectory := ""
	pdf := fpdf.New(orientation, unit, paperSize, fontDirectory)

	pdf.AddPage()

	columns := layout.GetColumns()
	rows := layout.GetRows()
	for i := 0; i < columns; i++ {
		for j := 0; j < rows; j++ {
			rectangle := layout.GetRectangle(i, j)
			generateLabel(pdf, &rectangle)

		}
	}

	return pdf.OutputFileAndClose(filename)
}

func isA4Size(layout *ValidatedLayoutSettings) bool {
	return layout.paperWidth == 210.0 && layout.paperHeight == 297.0
}

func generateLabel(pdf *fpdf.Fpdf, rectangle *Rectangle) {
	pdf.Rect(rectangle.Left, rectangle.Top, rectangle.Width, rectangle.Height, "D")
}
