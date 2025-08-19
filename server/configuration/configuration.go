package configuration

import (
	"encoding/json"
)

type Configuration struct {
	Log             *LogConfiguration             `json:"log"`
	LabelGeneration *LabelGenerationConfiguration `json:"labelGeneration"`
	Server          *ServerConfiguration          `json:"server"`
}

type LabelGenerationConfiguration struct {
	BarcodeWidth  int                `json:"barcodeWidth"`
	BarcodeHeight int                `json:"barcodeHeight"`
	Font          *FontConfiguration `json:"font"`
}

type FontConfiguration struct {
	Directory string `json:"directory"`
	Filename  string `json:"filename"`
	Family    string `json:"family"`
}

func (configuration *Configuration) String() string {
	bytes, err := json.MarshalIndent(configuration, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

type LogConfiguration struct {
	File             string `json:"file"`
	MaxSizeMegabytes int    `json:"maxSizeMegabytes"`
	MaxBackups       int    `json:"maxBackups"`
	MaxAgeDays       int    `json:"maxAgeDays"`
	Compression      bool   `json:"compression"`
}

type ServerConfiguration struct {
	Port                        int    `json:"port"`
	GinMode                     string `json:"ginMode"` // GinMode can be "debug", "release", or "test"
	ExpiredSessionPruneInterval int    `json:"expiredSessionPruneInterval"`
	CookieDomain                string `json:"cookieDomain"`
	HTMLPath                    string `json:"htmlPath"`
	Swagger                     bool   `json:"swagger"`
}
