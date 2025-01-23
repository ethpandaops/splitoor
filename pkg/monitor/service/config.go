package service

import (
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/validator"
)

type Config struct {
	Split     *split.Config     `yaml:"split"`
	Validator *validator.Config `yaml:"validator"`
}

func (c *Config) Validate() error {
	if c.Split != nil {
		if err := c.Split.Validate(); err != nil {
			return err
		}
	}

	if c.Validator != nil {
		if err := c.Validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}
