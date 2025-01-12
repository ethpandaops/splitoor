package split

import (
	"fmt"

	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/account"
)

const ServiceType = "split"

type Config struct {
	Splits []Split `yaml:"splits"`
}

type Split struct {
	Name       string            `yaml:"name"`
	Address    string            `yaml:"address"`
	Contract   string            `yaml:"contract"`
	Accounts   []*account.Config `yaml:"accounts"`
	Controller Controller        `yaml:"controller"`
}

type Controller struct {
	Type   string           `yaml:"type"`
	Config ControllerConfig `yaml:"config"`
}

type ControllerConfig struct {
	Address       string `yaml:"address"`
	MinSignatures int    `yaml:"minSignatures,omitempty"`
}

func (c *Config) Validate() error {
	for i := range c.Splits {
		if err := c.Splits[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Split) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	if c.Address == "" {
		return fmt.Errorf("address is required")
	}

	if len(c.Accounts) == 0 {
		return fmt.Errorf("accounts are required")
	}

	totalAllocation := int64(0)

	for _, a := range c.Accounts {
		if err := a.Validate(); err != nil {
			return err
		}

		totalAllocation += a.Allocation
	}

	if totalAllocation != 1000000 {
		return fmt.Errorf("total allocation must be 1000000 (100%%)")
	}

	if c.Controller.Type == "" {
		return fmt.Errorf("controller type is required")
	}

	if c.Controller.Config.Address == "" {
		return fmt.Errorf("controller address is required")
	}

	return nil
}
