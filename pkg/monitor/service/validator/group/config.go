package group

import "fmt"

type Config struct {
	Name    string   `yaml:"name"`
	Pubkeys []string `yaml:"pubkeys"`
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	return nil
}
