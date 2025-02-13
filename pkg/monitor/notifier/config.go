package notifier

import "github.com/ethpandaops/splitoor/pkg/monitor/notifier/source"

type Config struct {
	Sources []source.Config `yaml:"sources"`
	Docs    *string         `yaml:"docs"`
}

func (c *Config) Validate() error {
	return nil
}
