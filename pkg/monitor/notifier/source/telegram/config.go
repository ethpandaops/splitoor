package telegram

import (
	"errors"
)

type Config struct {
	BotToken string `yaml:"botToken" default:""`
	ChatID   string `yaml:"chatId" default:""`
}

func (c *Config) Validate() error {
	if c.BotToken == "" {
		return errors.New("telegram bot token is required")
	}

	if c.ChatID == "" {
		return errors.New("telegram chat id is required")
	}

	return nil
}
