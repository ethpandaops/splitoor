package group

import (
	"testing"

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
			name: "valid config - full config",
			config: &Config{
				Name:            "test_group",
				Address:         "0x123",
				RecoveryAddress: "0x789",
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
			expectError: false,
		},
		{
			name: "invalid config - one account",
			config: &Config{
				Name:            "test_group",
				Address:         "0x123",
				RecoveryAddress: "0x789",
				Accounts: []*account.Config{
					{
						Name:       "account1",
						Address:    "0x456",
						Allocation: 1000000,
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid config - no accounts",
			config: &Config{
				Name:            "test_group",
				Address:         "0x123",
				RecoveryAddress: "0x789",
				Accounts:        []*account.Config{},
			},
			expectError: true,
		},
		{
			name: "invalid config - bad account allocations",
			config: &Config{
				Name:            "test_group",
				Address:         "0x123",
				RecoveryAddress: "0x789",
				Accounts: []*account.Config{
					{
						Name:       "account1",
						Address:    "0x456",
						Allocation: 1000000,
					},
					{
						Name:       "account1",
						Address:    "0x456",
						Allocation: 1000000,
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid config - empty name",
			config: &Config{
				Name:            "",
				Address:         "0x123",
				RecoveryAddress: "0x789",
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
			},
			expectError: true,
		},
		{
			name: "invalid config - empty address",
			config: &Config{
				Name:            "test_group",
				Address:         "",
				RecoveryAddress: "0x789",
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
			},
			expectError: true,
		},
		{
			name: "invalid config - empty recovery address",
			config: &Config{
				Name:            "test_group",
				Address:         "0x123",
				RecoveryAddress: "",
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
			},
			expectError: true,
		},
		{
			name: "invalid config - invalid total allocation",
			config: &Config{
				Name:            "test_group",
				Address:         "0x123",
				RecoveryAddress: "0x789",
				Accounts: []*account.Config{
					{
						Name:       "account1",
						Address:    "0x456",
						Allocation: 500000,
					},
					{
						Name:       "account2",
						Address:    "0x789",
						Allocation: 400000,
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid config - invalid account",
			config: &Config{
				Name:            "test_group",
				Address:         "0x123",
				RecoveryAddress: "0x789",
				Accounts: []*account.Config{
					{
						Name:       "",
						Address:    "0x456",
						Allocation: 1000000,
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid config - invalid controller",
			config: &Config{
				Name:            "test_group",
				Address:         "0x123",
				RecoveryAddress: "0x789",
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
					ControllerType: controller.ControllerTypeUnknown,
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
