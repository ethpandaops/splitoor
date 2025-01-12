package safe

import (
	"context"

	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/sirupsen/logrus"
)

const ControllerType = "safe"

type Safe struct {
	log          logrus.FieldLogger
	name         string
	config       *Config
	ethereumPool *ethereum.Pool
}

func New(ctx context.Context, log logrus.FieldLogger, name string, config *Config, ethereumPool *ethereum.Pool) (*Safe, error) {
	return &Safe{
		log:          log,
		name:         name,
		config:       config,
		ethereumPool: ethereumPool,
	}, nil
}

func (c *Safe) Start(ctx context.Context) error {
	return nil
}

func (c *Safe) Stop(ctx context.Context) error {
	return nil
}

func (c *Safe) Type() string {
	return ControllerType
}

func (c *Safe) Name() string {
	return c.name
}
