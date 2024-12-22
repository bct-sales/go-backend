package barcode

import (
	"image"

	bclib "github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
)

func GenerateBarcode(data string, width int, height int) (image.Image, error) {
	bc, err := code128.Encode(data)

	if err != nil {
		return nil, err
	}

	scaledBarcode, err := bclib.Scale(bc, width, height)

	if err != nil {
		return nil, err
	}

	return scaledBarcode, nil
}
