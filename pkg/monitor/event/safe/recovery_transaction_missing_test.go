package safe_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/ethpandaops/splitoor/pkg/monitor/event/safe"
)

func TestRecoveryTransactionMissing(t *testing.T) {
	tests := []struct {
		name        string
		timestamp   time.Time
		monitor     string
		group       string
		safeAddress string
		wantTitle   string
		wantDesc    string
	}{
		{
			name:        "basic event",
			timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:     "test_monitor",
			group:       "test_group",
			safeAddress: "0x123",
			wantTitle:   "[test_monitor] Safe account has no recovery transaction queued",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Safe Account: 0x123`,
		},
		{
			name:        "special characters",
			timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:     "test!@#",
			group:       "test$%^",
			safeAddress: "0x123&*()",
			wantTitle:   "[test!@#] Safe account has no recovery transaction queued",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test!@#
Group: test$%^
Safe Account: 0x123&*()`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := safe.NewRecoveryTransactionMissing(
				tt.timestamp,
				tt.monitor,
				tt.group,
				tt.safeAddress,
			)

			// Verify it implements Event interface
			var _ event.Event = evt

			// Test type constant
			assert.Equal(t, safe.RecoveryTransactionMissingType, evt.GetType())

			// Test getters
			assert.Equal(t, tt.monitor, evt.GetMonitor())
			assert.Equal(t, tt.group, evt.GetGroup())
			assert.Equal(t, tt.wantTitle, evt.GetTitle(true, true))
			assert.Equal(t, tt.wantDesc, evt.GetDescriptionText(true, true))

			// Test fields
			assert.Equal(t, tt.timestamp, evt.Timestamp)
			assert.Equal(t, tt.safeAddress, evt.SafeAddress)
		})
	}
}
