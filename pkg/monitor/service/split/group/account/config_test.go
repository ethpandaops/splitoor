package account

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
			name: "valid config - full allocation",
			config: &Config{
				Name:       "test_account",
				Address:    "0x123",
				Allocation: 999999,
				Monitor:    true,
			},
			expectError: false,
		},
		{
			name: "valid config - partial allocation",
			config: &Config{
				Name:       "test_account",
				Address:    "0x123",
				Allocation: 500000,
				Monitor:    false,
			},
			expectError: false,
		},
		{
			name: "invalid config - empty name",
			config: &Config{
				Name:       "",
				Address:    "0x123",
				Allocation: 999999,
			},
			expectError: true,
		},
		{
			name: "invalid config - empty address",
			config: &Config{
				Name:       "test_account",
				Address:    "",
				Allocation: 999999,
			},
			expectError: true,
		},
		{
			name: "invalid config - zero allocation",
			config: &Config{
				Name:       "test_account",
				Address:    "0x123",
				Allocation: 0,
			},
			expectError: true,
		},
		{
			name: "invalid config - allocation too high",
			config: &Config{
				Name:       "test_account",
				Address:    "0x123",
				Allocation: 1000000000,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
		})
	}
}
