package safe

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
			name: "valid config - with recovery",
			config: &Config{
				Address:         "0x123",
				MinSignatures:   2,
				RecoveryAddress: "0x456",
			},
			expectError: false,
		},
		{
			name: "valid config - without recovery",
			config: &Config{
				Address:       "0x123",
				MinSignatures: 2,
			},
			expectError: false,
		},
		{
			name: "invalid config - empty address",
			config: &Config{
				Address:       "",
				MinSignatures: 2,
			},
			expectError: true,
		},
		{
			name: "invalid config - zero min signatures",
			config: &Config{
				Address:       "0x123",
				MinSignatures: 0,
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
