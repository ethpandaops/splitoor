package group

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
			name: "valid config - with name and pubkeys",
			config: &Config{
				Name:    "test_group",
				Pubkeys: []string{"0x123", "0x456"},
			},
			expectError: false,
		},
		{
			name: "valid config - with name and empty pubkeys",
			config: &Config{
				Name:    "test_group",
				Pubkeys: []string{},
			},
			expectError: false,
		},
		{
			name: "invalid config - empty name",
			config: &Config{
				Name:    "",
				Pubkeys: []string{"0x123"},
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
