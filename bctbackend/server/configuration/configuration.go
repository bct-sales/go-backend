package configuration

type Configuration struct {
	FontDirectory string
	FontFilename  string
	FontFamily    string
	HTMLPath      string
	BarcodeWidth  int
	BarcodeHeight int
	Port          int
	GinMode       string // GinMode can be "debug", "release", or "test"
}
