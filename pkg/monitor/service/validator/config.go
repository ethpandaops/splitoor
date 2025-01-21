package validator

import (
	"fmt"
	"time"

	"github.com/ethpandaops/splitoor/pkg/monitor/service/validator/group"
)

const ServiceType = "validator"

type Config struct {
	CheckInterval time.Duration  `yaml:"checkInterval" default:"24h"`
	Groups        []group.Config `yaml:"groups"`
}

func (c *Config) Validate() error {
	for _, g := range c.Groups {
		if err := g.Validate(); err != nil {
			return fmt.Errorf("group config is invalid: %w", err)
		}
	}

	return nil
}
