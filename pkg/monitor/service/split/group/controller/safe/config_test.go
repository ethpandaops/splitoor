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
				Address:   "0x123",
				Threshold: 2,
				Signers:   []string{"0x456", "0x789"},
			},
			expectError: false,
		},
		{
			name: "valid config - without recovery",
			config: &Config{
				Address:   "0x123",
				Threshold: 2,
				Signers:   []string{"0x456", "0x789"},
			},
			expectError: false,
		},
		{
			name: "invalid config - empty address",
			config: &Config{
				Address:   "",
				Threshold: 2,
				Signers:   []string{"0x456", "0x789"},
			},
			expectError: true,
		},
		{
			name: "invalid config - zero min signatures",
			config: &Config{
				Address:   "0x123",
				Threshold: 0,
				Signers:   []string{"0x456", "0x789"},
			},
			expectError: true,
		},
		{
			name: "invalid config - empty signers",
			config: &Config{
				Address:   "0x123",
				Threshold: 2,
				Signers:   []string{},
			},
			expectError: true,
		},
		{
			name: "invalid config - nil signers",
			config: &Config{
				Address:   "0x123",
				Threshold: 2,
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
