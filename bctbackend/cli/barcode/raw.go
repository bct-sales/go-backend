package barcode

import (
	"bctbackend/barcode"
	"errors"
	"fmt"
	"image/png"
	"os"
)

func GenerateRawBarcode(data string, outputPath string, width int, height int) (r_err error) {
	image, err := barcode.GenerateBarcode(data, width, height)
	if err != nil {
		return fmt.Errorf("failed to generate barcode: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", outputPath, err)
	}
	defer func() { r_err = errors.Join(err, file.Close()) }()

	if err = png.Encode(file, image); err != nil {
		return fmt.Errorf("failed to encode image as PNG: %w", err)
	}

	return nil
}
