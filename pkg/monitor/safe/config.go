package safe

import "fmt"

type Config struct {
	Enabled  bool   `yaml:"enabled" default:"true"`
	Endpoint string `yaml:"endpoint" default:"https://safe-client.safe.global/v1"`
}

func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	return nil
}
