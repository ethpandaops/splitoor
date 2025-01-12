package discord

import (
	"context"

	"github.com/sirupsen/logrus"
)

const ControllerType = "eoa"

type Discord struct {
	log        logrus.FieldLogger
	name       string
	sourceName string
	config     *Config
}

func NewDiscord(ctx context.Context, log logrus.FieldLogger, monitor, sourceName string, config *Config) (*Discord, error) {
	return &Discord{
		log:        log,
		sourceName: sourceName,
		config:     config,
	}, nil
}

func (c *Discord) Start(ctx context.Context) error {
	return nil
}

func (c *Discord) Stop(ctx context.Context) error {
	return nil
}

func (c *Discord) Type() string {
	return ControllerType
}

func (c *Discord) Name() string {
	return c.name
}

func (c *Discord) Publish(ctx context.Context, msg string) error {
	c.log.WithFields(logrus.Fields{
		"message": msg,
	}).Info("Sending message to discord")

	return nil
}
