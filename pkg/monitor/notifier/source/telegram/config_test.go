package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				BotToken: "test-token",
				ChatID:   "123456789",
			},
			expectError: false,
		},
		{
			name: "empty bot token",
			config: &Config{
				ChatID: "123456789",
			},
			expectError: true,
		},
		{
			name: "empty chat id",
			config: &Config{
				BotToken: "test-token",
			},
			expectError: true,
		},
		{
			name:        "empty config",
			config:      &Config{},
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
