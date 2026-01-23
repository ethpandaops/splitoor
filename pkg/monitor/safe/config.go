package safe

import "fmt"

// ChainIDToName maps chain IDs to their Safe API chain names.
var ChainIDToName = map[string]string{
	"1":        "eth", // Ethereum Mainnet
	"11155111": "sep", // Sepolia
	"560048":   "hod", // Hoodi
}

// Config holds configuration for the Safe API client.
type Config struct {
	Enabled  bool   `yaml:"enabled" default:"true"`
	Endpoint string `yaml:"endpoint" default:"https://api.safe.global"`
	APIKey   string `yaml:"apiKey"`
}

// Validate validates the Safe configuration.
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	if c.APIKey == "" {
		return fmt.Errorf("apiKey is required for Transaction Service API")
	}

	return nil
}
