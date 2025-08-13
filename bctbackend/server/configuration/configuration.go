package configuration

type Configuration struct {
	LogFilename                 *string
	LabelGeneration             *LabelGenerationConfiguration
	HTMLPath                    string
	Port                        int
	GinMode                     string // GinMode can be "debug", "release", or "test"
	ExpiredSessionPruneInterval int
}

type LabelGenerationConfiguration struct {
	BarcodeWidth  int
	BarcodeHeight int
	Font          *FontConfiguration
}

type FontConfiguration struct {
	Directory string
	Filename  string
	Family    string
}
