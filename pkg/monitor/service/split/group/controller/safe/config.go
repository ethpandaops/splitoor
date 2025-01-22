package safe

import "fmt"

type Config struct {
	Address         string `yaml:"address"`
	MinSignatures   int    `yaml:"minSignatures"`
	RecoveryAddress string `yaml:"recoveryAddress"`
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	if c.Address == "" {
		return fmt.Errorf("address is required")
	}

	if c.MinSignatures == 0 {
		return fmt.Errorf("minSignatures is required")
	}

	return nil
}
