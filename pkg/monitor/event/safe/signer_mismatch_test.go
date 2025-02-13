package safe_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/ethpandaops/splitoor/pkg/monitor/event/safe"
)

func TestSignerMismatch(t *testing.T) {
	tests := []struct {
		name         string
		timestamp    time.Time
		monitor      string
		group        string
		safeAddress  string
		wantTitle    string
		wantDesc     string
		wantDescMD   string
		wantDescHTML string
	}{
		{
			name:        "basic event",
			timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:     "test_monitor",
			group:       "test_group",
			safeAddress: "0x123",
			wantTitle:   "[test_monitor] Safe account has unexpected owners",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Safe Account: 0x123`,
			wantDescMD: `**Timestamp:** 2024-01-01 12:00:00 UTC
**Monitor:** test_monitor
**Group:** test_group
**Safe Account:** ` + "`0x123`",
			wantDescHTML: `<p><strong>Timestamp:</strong> 2024-01-01 12:00:00 UTC</p><p><strong>Monitor:</strong> test_monitor</p><p><strong>Group:</strong> test_group</p><p><strong>Safe Account:</strong> 0x123</p>`,
		},
		{
			name:        "special characters",
			timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:     "test!@#",
			group:       "test$%^",
			safeAddress: "0x123&*()",
			wantTitle:   "[test!@#] Safe account has unexpected owners",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test!@#
Group: test$%^
Safe Account: 0x123&*()`,
			wantDescMD: `**Timestamp:** 2024-01-01 12:00:00 UTC
**Monitor:** test!@#
**Group:** test$%^
**Safe Account:** ` + "`0x123&*()`",
			wantDescHTML: `<p><strong>Timestamp:</strong> 2024-01-01 12:00:00 UTC</p><p><strong>Monitor:</strong> test!@#</p><p><strong>Group:</strong> test$%^</p><p><strong>Safe Account:</strong> 0x123&*()</p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := safe.NewSignerMismatch(
				tt.timestamp,
				tt.monitor,
				tt.group,
				tt.safeAddress,
			)

			// Verify it implements Event interface
			var _ event.Event = evt

			// Test type constant
			assert.Equal(t, safe.SignerMismatchType, evt.GetType())

			// Test getters
			assert.Equal(t, tt.monitor, evt.GetMonitor())
			assert.Equal(t, "safe", evt.GetGroup())
			assert.Equal(t, tt.wantTitle, evt.GetTitle(true, true))
			assert.Equal(t, tt.wantDesc, evt.GetDescriptionText(true, true))
			assert.Equal(t, tt.wantDescMD, evt.GetDescriptionMarkdown(true, true))
			assert.Equal(t, tt.wantDescHTML, evt.GetDescriptionHTML(true, true))

			// Test fields
			assert.Equal(t, tt.timestamp, evt.Timestamp)
			assert.Equal(t, tt.safeAddress, evt.SafeAddress)
		})
	}
}
