package eoa

import "fmt"

type Config struct {
	Address string `yaml:"address"`
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	if c.Address == "" {
		return fmt.Errorf("address is required")
	}

	return nil
}
