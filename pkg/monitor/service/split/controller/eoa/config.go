package eoa

import "errors"

type Config struct {
	Address string `yaml:"address"`
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("address is required")
	}

	return nil
}
