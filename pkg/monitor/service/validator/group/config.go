package group

import "errors"

type Config struct {
	Name    string   `yaml:"name"`
	Pubkeys []string `yaml:"pubkeys"`
}

func (c *Config) Validate() error {
	if c == nil {
		return nil
	}

	if c.Name == "" {
		return errors.New("name is required")
	}

	return nil
}
