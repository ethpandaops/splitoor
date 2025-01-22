package smtp_test

import (
	"testing"

	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/smtp"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *smtp.Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &smtp.Config{
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
			config: &smtp.Config{
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
			config: &smtp.Config{
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
			config: &smtp.Config{
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
			config: &smtp.Config{
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
