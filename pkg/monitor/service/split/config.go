package split

import (
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group"
)

const ServiceType = "split"

type Config struct {
	Groups []group.Config `yaml:"groups"`
}

func (c *Config) Validate() error {
	for _, g := range c.Groups {
		if err := g.Validate(); err != nil {
			return err
		}
	}

	return nil
}
