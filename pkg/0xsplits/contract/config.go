package contract

import (
	"errors"
)

type Config struct {
	From       string `yaml:"from"`
	PrivateKey string `yaml:"privateKey"`
	GasLimit   uint64 `yaml:"gaslimit" default:"2940439"`
}

func (c *Config) Validate() error {
	if c.PrivateKey == "" {
		return errors.New("private key is required")
	}

	if c.From == "" {
		return errors.New("from address is required")
	}

	return nil
}
