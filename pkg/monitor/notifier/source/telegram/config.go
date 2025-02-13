package telegram

import (
	"errors"
)

type Config struct {
	BotToken string `yaml:"botToken"`
	ChatID   string `yaml:"chatId"`
	ThreadID int64  `yaml:"threadId,omitempty"`
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
