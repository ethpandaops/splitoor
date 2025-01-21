package safe_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/ethpandaops/splitoor/pkg/monitor/event/safe"
)

func TestTransactionQueueExcess(t *testing.T) {
	tests := []struct {
		name        string
		timestamp   time.Time
		monitor     string
		group       string
		safeAddress string
		count       int
		wantTitle   string
		wantDesc    string
	}{
		{
			name:        "basic event",
			timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:     "test_monitor",
			group:       "test_group",
			safeAddress: "0x123",
			count:       5,
			wantTitle:   "[test_monitor] test_group safe has unexpected transactions in queue",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Safe Account: 0x123 (https://app.safe.global/home?safe=0x123)
Number of Transactions: 5 (https://app.safe.global/transactions/queue?safe=0x123)`,
		},
		{
			name:        "large queue",
			timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:     "test_monitor",
			group:       "test_group",
			safeAddress: "0x123",
			count:       100,
			wantTitle:   "[test_monitor] test_group safe has unexpected transactions in queue",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Safe Account: 0x123 (https://app.safe.global/home?safe=0x123)
Number of Transactions: 100 (https://app.safe.global/transactions/queue?safe=0x123)`,
		},
		{
			name:        "special characters",
			timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:     "test!@#",
			group:       "test$%^",
			safeAddress: "0x123&*()",
			count:       5,
			wantTitle:   "[test!@#] test$%^ safe has unexpected transactions in queue",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test!@#
Group: test$%^
Safe Account: 0x123&*() (https://app.safe.global/home?safe=0x123&*())
Number of Transactions: 5 (https://app.safe.global/transactions/queue?safe=0x123&*())`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := safe.NewTransactionQueueExcess(
				tt.timestamp,
				tt.monitor,
				tt.group,
				tt.safeAddress,
				tt.count,
			)

			// Verify it implements Event interface
			var _ event.Event = evt

			// Test type constant
			assert.Equal(t, safe.TransactionQueueExcessType, evt.GetType())

			// Test getters
			assert.Equal(t, tt.monitor, evt.GetMonitor())
			assert.Equal(t, tt.group, evt.GetGroup())
			assert.Equal(t, tt.wantTitle, evt.GetTitle())
			assert.Equal(t, tt.wantDesc, evt.GetDescription())

			// Test fields
			assert.Equal(t, tt.timestamp, evt.Timestamp)
			assert.Equal(t, tt.safeAddress, evt.SafeAddress)
			assert.Equal(t, tt.count, evt.NumTxs)
		})
	}
}
