package configuration

type Configuration struct {
	LogFilename                 *string
	Font                        *FontConfiguration
	HTMLPath                    string
	BarcodeWidth                int
	BarcodeHeight               int
	Port                        int
	GinMode                     string // GinMode can be "debug", "release", or "test"
	ExpiredSessionPruneInterval int
}

type FontConfiguration struct {
	Directory string
	Filename  string
	Family    string
}
