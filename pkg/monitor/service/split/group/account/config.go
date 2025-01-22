package account

import (
	"fmt"
)

type Config struct {
	Name       string `yaml:"name"`
	Address    string `yaml:"address"`
	Allocation uint32 `yaml:"allocation"`
	Monitor    bool   `yaml:"monitor" default:"false"`
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	if c.Address == "" {
		return fmt.Errorf("address is required")
	}

	if c.Allocation <= 0 {
		return fmt.Errorf("allocations must be greater than 0")
	}

	if c.Allocation > 999999 {
		return fmt.Errorf("allocation must be less than 999999 (99.9999%%)")
	}

	return nil
}
