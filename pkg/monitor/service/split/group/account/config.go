package account

import (
	"errors"
)

type Config struct {
	Name       string `yaml:"name"`
	Address    string `yaml:"address"`
	Allocation uint32 `yaml:"allocation"`
	Monitor    bool   `yaml:"monitor" default:"false"`
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is nil")
	}

	if c.Name == "" {
		return errors.New("name is required")
	}

	if c.Address == "" {
		return errors.New("address is required")
	}

	if c.Allocation <= 0 {
		return errors.New("allocations must be greater than 0")
	}

	if c.Allocation > 999999 {
		return errors.New("allocation must be less than 999999 (99.9999%%)")
	}

	return nil
}
