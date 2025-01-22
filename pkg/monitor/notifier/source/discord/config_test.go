package discord_test

import (
	"testing"

	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/discord"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *discord.Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &discord.Config{
				Webhook: "https://discord.com/api/webhooks/test",
			},
			expectError: false,
		},
		{
			name:        "empty webhook",
			config:      &discord.Config{},
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
