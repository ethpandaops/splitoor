package ethereum

import (
	"fmt"

	"github.com/ethpandaops/splitoor/pkg/ethereum/beacon"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
)

type Config struct {
	// Execution configuration
	Execution []*execution.Config `yaml:"execution"`
	// Beacon configuration
	Beacon []*beacon.Config `yaml:"beacon"`
}

func (c *Config) Validate() error {
	for i, execution := range c.Execution {
		if err := execution.Validate(); err != nil {
			return fmt.Errorf("invalid execution configuration at index %d: %w", i, err)
		}
	}

	for i, beacon := range c.Beacon {
		if err := beacon.Validate(); err != nil {
			return fmt.Errorf("invalid beacon configuration at index %d: %w", i, err)
		}
	}

	return nil
}
