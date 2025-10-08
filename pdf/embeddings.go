package pdf

import (
	"bytes"
	_ "embed"
)

//go:embed arial.ttf
var embeddedFont []byte

//go:embed charity.png
var charityImage []byte

//go:embed donation.png
var donationImage []byte

func convertByteArrayToBuffer(data []byte) *bytes.Buffer {
	return bytes.NewBuffer(data)
}

func DonationImageAsBuffer() *bytes.Buffer {
	return convertByteArrayToBuffer(donationImage)
}

func CharityImageAsBuffer() *bytes.Buffer {
	return convertByteArrayToBuffer(charityImage)
}
