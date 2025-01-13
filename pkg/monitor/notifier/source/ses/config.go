package ses

import "fmt"

type Config struct {
	From string   `yaml:"from"`
	To   []string `yaml:"to"`
}

func (c *Config) Validate() error {
	if c.From == "" {
		return fmt.Errorf("from is required")
	}

	if len(c.To) == 0 {
		return fmt.Errorf("to is required")
	}

	return nil
}
