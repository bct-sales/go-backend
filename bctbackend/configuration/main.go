package configuration

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	DatabasePath  string `yaml:"database_path"`
	FontDirectory string `yaml:"font_directory"`
}

type ConfigurationError struct {
	Message string
	Wrapped error
}

func (e *ConfigurationError) Error() string {
	return e.Message + ": " + e.Wrapped.Error()
}

func (e *ConfigurationError) Unwrap() error {
	return e.Wrapped
}

func LoadConfigurationFromFile(path string) (*Configuration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ConfigurationError{
			Message: "Failed to read configuration file",
			Wrapped: err,
		}
	}

	return LoadConfigurationFromBytes(data)
}

func LoadConfigurationFromBytes(data []byte) (*Configuration, error) {
	var config Configuration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, &ConfigurationError{
			Message: "Failed to parse configuration bytes",
			Wrapped: err,
		}
	}

	return &config, nil
}
