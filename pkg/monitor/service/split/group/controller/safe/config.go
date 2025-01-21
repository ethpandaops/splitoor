package safe

import "errors"

type Config struct {
	Address         string `yaml:"address"`
	MinSignatures   int    `yaml:"minSignatures"`
	RecoveryAddress string `yaml:"recoveryAddress"`
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("address is required")
	}

	if c.MinSignatures == 0 {
		return errors.New("minSignatures is required")
	}

	return nil
}
