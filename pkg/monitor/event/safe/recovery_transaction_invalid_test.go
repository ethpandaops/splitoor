package safe_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/ethpandaops/splitoor/pkg/monitor/event/safe"
)

func TestRecoveryTransactionInvalid(t *testing.T) {
	tests := []struct {
		name        string
		timestamp   time.Time
		monitor     string
		group       string
		safeAddress string
		txID        string
		reason      string
		wantTitle   string
		wantDesc    string
	}{
		{
			name:        "basic event",
			timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:     "test_monitor",
			group:       "test_group",
			safeAddress: "0x123",
			txID:        "tx_123",
			reason:      "invalid signature",
			wantTitle:   "[test_monitor] Safe account has invalid recovery transaction",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Safe Account: 0x123
Transaction ID: tx_123
Reason: invalid signature`,
		},
		{
			name:        "empty reason",
			timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:     "test_monitor",
			group:       "test_group",
			safeAddress: "0x123",
			txID:        "tx_123",
			reason:      "",
			wantTitle:   "[test_monitor] Safe account has invalid recovery transaction",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Safe Account: 0x123
Transaction ID: tx_123
Reason: `,
		},
		{
			name:        "special characters",
			timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:     "test!@#",
			group:       "test$%^",
			safeAddress: "0x123&*()",
			txID:        "tx_123!@#",
			reason:      "invalid!@#",
			wantTitle:   "[test!@#] Safe account has invalid recovery transaction",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test!@#
Group: test$%^
Safe Account: 0x123&*()
Transaction ID: tx_123!@#
Reason: invalid!@#`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := safe.NewRecoveryTransactionInvalid(
				tt.timestamp,
				tt.monitor,
				tt.group,
				tt.safeAddress,
				tt.txID,
				tt.reason,
			)

			// Verify it implements Event interface
			var _ event.Event = evt

			// Test type constant
			assert.Equal(t, safe.RecoveryTransactionInvalidType, evt.GetType())

			// Test getters
			assert.Equal(t, tt.monitor, evt.GetMonitor())
			assert.Equal(t, tt.group, evt.GetGroup())
			assert.Equal(t, tt.wantTitle, evt.GetTitle(true, true))
			assert.Equal(t, tt.wantDesc, evt.GetDescriptionText(true, true))

			// Test fields
			assert.Equal(t, tt.timestamp, evt.Timestamp)
			assert.Equal(t, tt.safeAddress, evt.SafeAddress)
			assert.Equal(t, tt.txID, evt.RecoveryTransactionID)
			assert.Equal(t, tt.reason, evt.Reason)
		})
	}
}
