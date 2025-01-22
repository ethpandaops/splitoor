package telegram_test

import (
	"testing"

	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/telegram"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *telegram.Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &telegram.Config{
				BotToken: "test-token",
				ChatID:   "123456789",
			},
			expectError: false,
		},
		{
			name: "empty bot token",
			config: &telegram.Config{
				ChatID: "123456789",
			},
			expectError: true,
		},
		{
			name: "empty chat id",
			config: &telegram.Config{
				BotToken: "test-token",
			},
			expectError: true,
		},
		{
			name:        "empty config",
			config:      &telegram.Config{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
