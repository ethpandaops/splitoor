package group

import (
	"fmt"

	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/account"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/controller"
)

type Config struct {
	Name       string            `yaml:"name"`
	Address    string            `yaml:"address"`
	Contract   *string           `yaml:"contract"`
	Accounts   []*account.Config `yaml:"accounts"`
	Controller controller.Config `yaml:"controller"`
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	if c.Address == "" {
		return fmt.Errorf("address is required")
	}

	totalAllocation := uint32(0)

	for _, a := range c.Accounts {
		if err := a.Validate(); err != nil {
			return err
		}

		totalAllocation += a.Allocation
	}

	if totalAllocation != 1000000 {
		return fmt.Errorf("total allocation must be 1000000 (100%%)")
	}

	if err := c.Controller.Validate(); err != nil {
		return err
	}

	return nil
}
