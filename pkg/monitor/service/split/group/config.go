package group

import (
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/account"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/controller"
	"github.com/pkg/errors"
)

type Config struct {
	Name            string            `yaml:"name"`
	Address         string            `yaml:"address"`
	RecoveryAddress string            `yaml:"recoveryAddress"`
	Contract        *string           `yaml:"contract"`
	Accounts        []*account.Config `yaml:"accounts"`
	Controller      controller.Config `yaml:"controller"`
}

func (c *Config) Validate() error {
	if c == nil {
		return nil
	}

	if c.Name == "" {
		return errors.New("name is required")
	}

	if c.Address == "" {
		return errors.New("address is required")
	}

	if c.RecoveryAddress == "" {
		return errors.New("recoveryAddress is required")
	}

	totalAllocation := uint32(0)

	for _, a := range c.Accounts {
		if err := a.Validate(); err != nil {
			return err
		}

		totalAllocation += a.Allocation
	}

	if totalAllocation != 1000000 {
		return errors.New("total allocation must be 1000000 (100%%)")
	}

	if err := c.Controller.Validate(); err != nil {
		return err
	}

	return nil
}
