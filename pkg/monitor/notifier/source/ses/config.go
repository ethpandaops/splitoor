package ses

import "errors"

type Config struct {
	From string   `yaml:"from"`
	To   []string `yaml:"to"`
}

func (c *Config) Validate() error {
	if c.From == "" {
		return errors.New("from is required")
	}

	if len(c.To) == 0 {
		return errors.New("to is required")
	}

	return nil
}
