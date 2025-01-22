package validator_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/ethpandaops/splitoor/pkg/monitor/event/validator"
)

func TestWithdrawalCredentials(t *testing.T) {
	tests := []struct {
		name      string
		timestamp time.Time
		code      int64
		pubkey    string
		group     string
		monitor   string
		wantTitle string
		wantDesc  string
	}{
		{
			name:      "basic event",
			timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			code:      0x00,
			pubkey:    "0x123",
			group:     "test_group",
			monitor:   "test_monitor",
			wantTitle: "[test_monitor] test_group validator has unexpected withdrawal credentials type",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Pubkey: 0x123
Code: 0x00`,
		},
		{
			name:      "different code",
			timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			code:      0x01,
			pubkey:    "0x123",
			group:     "test_group",
			monitor:   "test_monitor",
			wantTitle: "[test_monitor] test_group validator has unexpected withdrawal credentials type",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Pubkey: 0x123
Code: 0x01`,
		},
		{
			name:      "large code",
			timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			code:      0xff,
			pubkey:    "0x123",
			group:     "test_group",
			monitor:   "test_monitor",
			wantTitle: "[test_monitor] test_group validator has unexpected withdrawal credentials type",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Pubkey: 0x123
Code: 0xff`,
		},
		{
			name:      "special characters",
			timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			code:      0x00,
			pubkey:    "0x123!@#",
			group:     "test$%^",
			monitor:   "test&*()",
			wantTitle: "[test&*()] test$%^ validator has unexpected withdrawal credentials type",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test&*()
Group: test$%^
Pubkey: 0x123!@#
Code: 0x00`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := validator.NewWithdrawalCredentials(
				tt.timestamp,
				tt.code,
				tt.pubkey,
				tt.group,
				tt.monitor,
			)

			// Verify it implements Event interface
			var _ event.Event = evt

			// Test type constant
			assert.Equal(t, validator.WithdrawalCredentialsType, evt.GetType())

			// Test getters
			assert.Equal(t, tt.monitor, evt.GetMonitor())
			assert.Equal(t, tt.group, evt.GetGroup())
			assert.Equal(t, tt.wantTitle, evt.GetTitle())
			assert.Equal(t, tt.wantDesc, evt.GetDescription())

			// Test fields
			assert.Equal(t, tt.timestamp, evt.Timestamp)
			assert.Equal(t, tt.pubkey, evt.Pubkey)
			assert.Equal(t, tt.code, evt.Code)
		})
	}
}
