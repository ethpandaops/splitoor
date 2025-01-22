package split

import (
	"testing"

	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/account"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/controller"
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
				Groups: []group.Config{},
			},
			expectError: false,
		},
		{
			name: "valid config - with groups",
			config: &Config{
				Groups: []group.Config{
					{
						Name:    "test_group",
						Address: "0x123",
						Accounts: []*account.Config{
							{
								Name:       "account1",
								Address:    "0x456",
								Allocation: 999999,
							},
							{
								Name:       "account1",
								Address:    "0x456",
								Allocation: 1,
							},
						},
						Controller: controller.Config{
							ControllerType: controller.ControllerTypeEOA,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid config - invalid group",
			config: &Config{
				Groups: []group.Config{
					{
						Name:    "", // Empty name should fail validation
						Address: "0x123",
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
