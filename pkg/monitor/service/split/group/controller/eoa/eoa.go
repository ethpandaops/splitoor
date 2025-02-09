package eoa

import (
	"context"

	"github.com/sirupsen/logrus"
)

const ControllerType = "eoa"

type EOA struct {
	log     logrus.FieldLogger
	name    string
	address string
}

func New(ctx context.Context, log logrus.FieldLogger, name string, config *Config) (*EOA, error) {
	return &EOA{
		log:     log,
		name:    name,
		address: config.Address,
	}, nil
}

func (c *EOA) Start(ctx context.Context) error {
	return nil
}

func (c *EOA) Stop(ctx context.Context) error {
	return nil
}

func (c *EOA) Type() string {
	return ControllerType
}

func (c *EOA) Name() string {
	return c.name
}

func (c *EOA) Address() string {
	return c.address
}
