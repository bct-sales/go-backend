package barcode

import (
	"image/png"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
)

func GenerateRawBarcode(data string, outputPath string, width int, height int) error {
	bc, err := code128.Encode(data)

	if err != nil {
		return err
	}

	scaledBarcode, err := barcode.Scale(bc, width, height)

	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)

	if err != nil {
		return err
	}

	defer file.Close()

	png.Encode(file, scaledBarcode)

	return nil
}
