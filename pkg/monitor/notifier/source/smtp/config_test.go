package smtp

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
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
				From:     "test@example.com",
				To:       []string{"recipient@example.com"},
			},
			expectError: false,
		},
		{
			name: "empty host",
			config: &Config{
				Port:     587,
				Username: "test@example.com",
				Password: "password",
				From:     "test@example.com",
				To:       []string{"recipient@example.com"},
			},
			expectError: true,
		},
		{
			name: "empty from",
			config: &Config{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
				To:       []string{"recipient@example.com"},
			},
			expectError: true,
		},
		{
			name: "empty recipients",
			config: &Config{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
				From:     "test@example.com",
			},
			expectError: true,
		},
		{
			name: "invalid port",
			config: &Config{
				Host:     "smtp.example.com",
				Port:     0,
				Username: "test@example.com",
				Password: "password",
				From:     "test@example.com",
				To:       []string{"recipient@example.com"},
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
