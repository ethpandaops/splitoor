package discord

import "errors"

type Config struct {
	Webhook string `yaml:"webhook"`
}

func (c *Config) Validate() error {
	if c.Webhook == "" {
		return errors.New("webhook is required")
	}

	return nil
}
