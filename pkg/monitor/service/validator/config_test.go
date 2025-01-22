package validator

import (
	"testing"
	"time"

	"github.com/ethpandaops/splitoor/pkg/monitor/service/validator/group"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config - empty groups",
			config: &Config{
				CheckInterval: 24 * time.Hour,
				Groups:        []group.Config{},
			},
			expectError: false,
		},
		{
			name: "valid config - with groups",
			config: &Config{
				CheckInterval: 24 * time.Hour,
				Groups: []group.Config{
					{
						Name:    "test_group",
						Pubkeys: []string{"0x123"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid config - invalid group",
			config: &Config{
				CheckInterval: 24 * time.Hour,
				Groups: []group.Config{
					{
						Name:    "", // Empty name should fail validation
						Pubkeys: []string{"0x123"},
					},
				},
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
