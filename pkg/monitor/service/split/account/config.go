package account

import (
	"fmt"
)

type Config struct {
	Name       string `yaml:"name"`
	Address    string `yaml:"address"`
	Allocation int64  `yaml:"allocation"`
}

func (c *Config) Validate() error {
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
