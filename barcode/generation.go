package barcode

import (
	"fmt"
	"image"
	"log/slog"

	bclib "github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
)

func GenerateBarcode(data string, width int, height int) (image.Image, error) {
	slog.Debug("Generating barcode", "data", data)
	barcode, err := code128.Encode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encode barcode: %w", err)
	}

	slog.Debug("Scaling barcode", "width", width, "height", height)
	scaledBarcode, err := bclib.Scale(barcode, width, height)
	if err != nil {
		return nil, fmt.Errorf("failed to scale barcode: %w", err)
	}

	return scaledBarcode, nil
}
