package monitor

import (
	"fmt"

	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/beaconchain"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	"github.com/ethpandaops/splitoor/pkg/monitor/safe"
	"github.com/ethpandaops/splitoor/pkg/monitor/service"
)

type Config struct {
	// MetricsAddr is the address to listen on for metrics.
	MetricsAddr string `yaml:"metricsAddr" default:":9090"`
	// PProfAddr is the address to listen on for pprof.
	PProfAddr *string `yaml:"pprofAddr"`
	// LoggingLevel is the logging level to use.
	LoggingLevel string `yaml:"logging" default:"info"`
	// Name is the name of the monitor.
	Name string `yaml:"name"`
	// Services is the list of services to run.
	Services service.Config `yaml:"services"`
	// Ethereum is the ethereum network configuration.
	Ethereum ethereum.Config `yaml:"ethereum"`
	// Notifier is the list of notifiers to run.
	Notifier notifier.Config `yaml:"notifier"`
	// Beaconchain is the beaconchain configuration.
	Beaconchain beaconchain.Config `yaml:"beaconchain"`
	// Safe is the safe configuration.
	Safe safe.Config `yaml:"safe"`
}

func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	if err := c.Services.Validate(); err != nil {
		return err
	}

	if err := c.Ethereum.Validate(); err != nil {
		return err
	}

	if err := c.Beaconchain.Validate(); err != nil {
		return err
	}

	if err := c.Safe.Validate(); err != nil {
		return err
	}

	return nil
}
