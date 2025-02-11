package safe

import "fmt"

type Config struct {
	Enabled  bool     `yaml:"enabled" default:"true"`
	Endpoint string   `yaml:"endpoint" default:"https://safe-client.safe.global"`
	Signers  []string `yaml:"signers"`
}

func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	return nil
}
