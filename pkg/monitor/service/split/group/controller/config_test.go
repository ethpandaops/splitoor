package controller

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
			name: "valid config - EOA controller",
			config: &Config{
				ControllerType: ControllerTypeEOA,
			},
			expectError: false,
		},
		{
			name: "valid config - Safe controller",
			config: &Config{
				ControllerType: ControllerTypeSafe,
			},
			expectError: false,
		},
		{
			name: "invalid config - unknown controller type",
			config: &Config{
				ControllerType: ControllerTypeUnknown,
			},
			expectError: true,
		},
		{
			name: "invalid config - empty controller type",
			config: &Config{
				ControllerType: "",
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
