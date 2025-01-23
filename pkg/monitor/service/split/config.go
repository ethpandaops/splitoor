package split

import (
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group"
)

const ServiceType = "split"

type Config struct {
	Groups []group.Config `yaml:"groups"`
}

func (c *Config) Validate() error {
	if c == nil {
		return nil
	}

	for _, g := range c.Groups {
		if err := g.Validate(); err != nil {
			return err
		}
	}

	return nil
}
