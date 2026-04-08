package group

import (
	"testing"

	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/account"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/controller"
	"github.com/stretchr/testify/assert"
)

// validBaseConfig returns a valid Config that individual tests can modify.
func validBaseConfig() *Config {
	return &Config{
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
				Name:       "account2",
				Address:    "0x456",
				Allocation: 1,
			},
		},
		Controller: controller.Config{
			ControllerType: controller.ControllerTypeEOA,
		},
	}
}

func TestConfigValidate_Valid(t *testing.T) {
	cfg := validBaseConfig()
	assert.NoError(t, cfg.Validate())
}

func TestConfigValidate_OneAccount(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Accounts = []*account.Config{
		{
			Name:       "account1",
			Address:    "0x456",
			Allocation: 1000000,
		},
	}

	assert.Error(t, cfg.Validate())
}

func TestConfigValidate_NoAccounts(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Accounts = []*account.Config{}

	assert.Error(t, cfg.Validate())
}

func TestConfigValidate_BadAllocations(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Accounts = []*account.Config{
		{
			Name:       "account1",
			Address:    "0x456",
			Allocation: 1000000,
		},
		{
			Name:       "account2",
			Address:    "0x456",
			Allocation: 1000000,
		},
	}

	assert.Error(t, cfg.Validate())
}

func TestConfigValidate_EmptyName(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Name = ""

	assert.Error(t, cfg.Validate())
}

func TestConfigValidate_EmptyAddress(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Address = ""

	assert.Error(t, cfg.Validate())
}

func TestConfigValidate_EmptyRecoveryAddress(t *testing.T) {
	cfg := validBaseConfig()
	cfg.RecoveryAddress = ""

	assert.Error(t, cfg.Validate())
}

func TestConfigValidate_InvalidTotalAllocation(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Accounts = []*account.Config{
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
	}

	assert.Error(t, cfg.Validate())
}

func TestConfigValidate_InvalidAccount(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Accounts = []*account.Config{
		{
			Name:       "",
			Address:    "0x456",
			Allocation: 1000000,
		},
	}

	assert.Error(t, cfg.Validate())
}

func TestConfigValidate_InvalidController(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Controller = controller.Config{
		ControllerType: controller.ControllerTypeUnknown,
	}

	assert.Error(t, cfg.Validate())
}
