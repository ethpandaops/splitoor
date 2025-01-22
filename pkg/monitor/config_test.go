package monitor_test

import (
	"testing"

	"github.com/ethpandaops/splitoor/pkg/ethereum"
	mon "github.com/ethpandaops/splitoor/pkg/monitor"
	"github.com/ethpandaops/splitoor/pkg/monitor/beaconchain"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	"github.com/ethpandaops/splitoor/pkg/monitor/safe"
	"github.com/ethpandaops/splitoor/pkg/monitor/service"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *mon.Config
		wantErr bool
	}{
		{
			name: "valid minimal config",
			config: &mon.Config{
				Name:        "monitor-1",
				MetricsAddr: ":9090",
				Services:    service.Config{},
				Ethereum:    ethereum.Config{},
				Notifier:    notifier.Config{},
				Beaconchain: beaconchain.Config{
					Enabled: false,
				},
				Safe: safe.Config{
					Enabled: false,
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: &mon.Config{
				MetricsAddr: ":9090",
				Services:    service.Config{},
				Ethereum:    ethereum.Config{},
				Notifier:    notifier.Config{},
				Beaconchain: beaconchain.Config{},
				Safe:        safe.Config{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
