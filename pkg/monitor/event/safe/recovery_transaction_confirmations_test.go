package safe_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/ethpandaops/splitoor/pkg/monitor/event/safe"
)

func TestRecoveryTransactionConfirmations(t *testing.T) {
	tests := []struct {
		name                  string
		timestamp             time.Time
		monitor               string
		group                 string
		safeAddress           string
		txID                  string
		numConfirmations      int
		expectedConfirmations int
		wantTitle             string
		wantDesc              string
	}{
		{
			name:                  "basic event",
			timestamp:             time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:               "test_monitor",
			group:                 "test_group",
			safeAddress:           "0x123",
			txID:                  "tx_123",
			numConfirmations:      1,
			expectedConfirmations: 3,
			wantTitle:             "[test_monitor] test_group safe account has a recovery transaction with incorrect number of confirmations",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Safe Account: 0x123 (https://app.safe.global/home?safe=0x123)
Recovery Transaction: tx_123 (https://app.safe.global/transactions/queue?safe=0x123)
Current Confirmations: 1
Expected Confirmations: 3`,
		},
		{
			name:                  "all confirmations submitted",
			timestamp:             time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:               "test_monitor",
			group:                 "test_group",
			safeAddress:           "0x123",
			txID:                  "tx_123",
			numConfirmations:      2,
			expectedConfirmations: 2,
			wantTitle:             "[test_monitor] test_group safe account has a recovery transaction with incorrect number of confirmations",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Safe Account: 0x123 (https://app.safe.global/home?safe=0x123)
Recovery Transaction: tx_123 (https://app.safe.global/transactions/queue?safe=0x123)
Current Confirmations: 2
Expected Confirmations: 2`,
		},
		{
			name:                  "zero confirmations",
			timestamp:             time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:               "test_monitor",
			group:                 "test_group",
			safeAddress:           "0x123",
			txID:                  "tx_123",
			numConfirmations:      0,
			expectedConfirmations: 2,
			wantTitle:             "[test_monitor] test_group safe account has a recovery transaction with incorrect number of confirmations",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Safe Account: 0x123 (https://app.safe.global/home?safe=0x123)
Recovery Transaction: tx_123 (https://app.safe.global/transactions/queue?safe=0x123)
Current Confirmations: 0
Expected Confirmations: 2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := safe.NewRecoveryTransactionConfirmations(
				tt.timestamp,
				tt.monitor,
				tt.group,
				tt.safeAddress,
				tt.txID,
				tt.numConfirmations,
				tt.expectedConfirmations,
			)

			// Verify it implements Event interface
			var _ event.Event = evt

			// Test type constant
			assert.Equal(t, safe.RecoveryTransactionConfirmationsType, evt.GetType())

			// Test getters
			assert.Equal(t, tt.monitor, evt.GetMonitor())
			assert.Equal(t, tt.group, evt.GetGroup())
			assert.Equal(t, tt.wantTitle, evt.GetTitle())
			assert.Equal(t, tt.wantDesc, evt.GetDescription())

			// Test fields
			assert.Equal(t, tt.timestamp, evt.Timestamp)
			assert.Equal(t, tt.safeAddress, evt.SafeAddress)
			assert.Equal(t, tt.txID, evt.RecoveryTransactionID)
			assert.Equal(t, tt.numConfirmations, evt.NumConfirmations)
			assert.Equal(t, tt.expectedConfirmations, evt.ExpectedConfirmations)
		})
	}
}
