package safe

import "fmt"

type Config struct {
	Address   string   `yaml:"address"`
	Signers   []string `yaml:"signers"`
	Threshold int      `yaml:"threshold"`
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	if c.Address == "" {
		return fmt.Errorf("address is required")
	}

	if c.Threshold == 0 {
		return fmt.Errorf("threshold is required")
	}

	if len(c.Signers) == 0 {
		return fmt.Errorf("signers is required")
	}

	return nil
}
