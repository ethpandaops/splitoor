package validator

import (
	"fmt"
	"time"

	"github.com/ethpandaops/splitoor/pkg/monitor/beaconchain"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/validator/group"
)

const ServiceType = "validator"

type Config struct {
	Beaconchain   beaconchain.Config `yaml:"beaconchain"`
	CheckInterval time.Duration      `yaml:"checkInterval" default:"24h"`
	Groups        []group.Config     `yaml:"groups"`
}

func (c *Config) Validate() error {
	for _, g := range c.Groups {
		if err := g.Validate(); err != nil {
			return fmt.Errorf("group config is invalid: %w", err)
		}
	}

	if c.Beaconchain.APIKey != "" {
		if err := c.Beaconchain.Validate(); err != nil {
			return fmt.Errorf("beaconchain config is invalid: %w", err)
		}
	}

	return nil
}
