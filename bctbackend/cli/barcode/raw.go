package barcode

import (
	"bctbackend/barcode"
	"errors"
	"image/png"
	"os"
)

func GenerateRawBarcode(data string, outputPath string, width int, height int) (r_err error) {
	image, err := barcode.GenerateBarcode(data, width, height)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(err, file.Close()) }()

	if err = png.Encode(file, image); err != nil {
		return err
	}

	return nil
}
