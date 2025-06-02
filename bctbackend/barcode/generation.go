package barcode

import (
	"fmt"
	"image"

	bclib "github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
)

func GenerateBarcode(data string, width int, height int) (image.Image, error) {
	barcode, err := code128.Encode(data)

	if err != nil {
		return nil, err
	}

	scaledBarcode, err := bclib.Scale(barcode, width, height)
	if err != nil {
		return nil, fmt.Errorf("failed to generate barcode: %w", err)
	}

	return scaledBarcode, nil
}
