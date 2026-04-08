package safe

import "errors"

type Config struct {
	Address   string   `yaml:"address"`
	Signers   []string `yaml:"signers"`
	Threshold int      `yaml:"threshold"`
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is nil")
	}

	if c.Address == "" {
		return errors.New("address is required")
	}

	if c.Threshold == 0 {
		return errors.New("threshold is required")
	}

	if len(c.Signers) == 0 {
		return errors.New("signers is required")
	}

	return nil
}
