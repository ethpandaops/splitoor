package beacon

import "github.com/pkg/errors"

type Config struct {
	// The address of the Beacon node to connect to
	NodeAddress string `yaml:"nodeAddress"`
	// NodeHeaders is a map of headers to send to the beacon node.
	NodeHeaders map[string]string `yaml:"nodeHeaders"`
	// Name is the name of the beacon node
	Name string `yaml:"name"`
}

func (c *Config) Validate() error {
	if c.NodeAddress == "" {
		return errors.New("beaconNodeAddress is required")
	}

	return nil
}
