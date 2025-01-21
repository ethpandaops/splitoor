package ses

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
				From: "test@example.com",
				To:   []string{"recipient@example.com"},
			},
			expectError: false,
		},
		{
			name: "empty from",
			config: &Config{
				To: []string{"recipient@example.com"},
			},
			expectError: true,
		},
		{
			name: "empty recipients",
			config: &Config{
				From: "test@example.com",
			},
			expectError: true,
		},
		{
			name: "empty recipient list",
			config: &Config{
				From: "test@example.com",
				To:   []string{},
			},
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
