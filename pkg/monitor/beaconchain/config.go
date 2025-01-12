package beaconchain

import (
	"fmt"
	"time"
)

type Config struct {
	Enabled              bool          `yaml:"enabled" default:"true"`
	APIKey               string        `yaml:"apiKey"`
	BatchSize            int           `yaml:"batchSize" default:"100"`
	MaxRequestsPerMinute int           `yaml:"maxRequestsPerMinute" default:"10"`
	Endpoint             string        `yaml:"endpoint" default:"https://beaconcha.in"`
	CheckInterval        time.Duration `yaml:"checkInterval" default:"24h"`
}

func (c *Config) Validate() error {
	if c.BatchSize <= 0 {
		return fmt.Errorf("batch size must be greater than 0")
	}

	if c.MaxRequestsPerMinute <= 0 {
		return fmt.Errorf("max requests per hour must be greater than 0")
	}

	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	return nil
}
