package split_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/ethpandaops/splitoor/pkg/monitor/event/split"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name         string
		timestamp    time.Time
		monitor      string
		group        string
		splitAddress string
		expectedHash string
		actualHash   string
		wantTitle    string
		wantDesc     string
	}{
		{
			name:         "basic event",
			timestamp:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:      "test_monitor",
			group:        "test_group",
			splitAddress: "0x123",
			expectedHash: "0x456",
			actualHash:   "0x789",
			wantTitle:    "[test_monitor] Split hash has changed",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Split Address: 0x123
Expected Hash: 0x456
Actual Hash: 0x789`,
		},
		{
			name:         "same hash",
			timestamp:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:      "test_monitor",
			group:        "test_group",
			splitAddress: "0x123",
			expectedHash: "0x456",
			actualHash:   "0x456",
			wantTitle:    "[test_monitor] Split hash has changed",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Split Address: 0x123
Expected Hash: 0x456
Actual Hash: 0x456`,
		},
		{
			name:         "special characters",
			timestamp:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:      "test!@#",
			group:        "test$%^",
			splitAddress: "0x123&*()",
			expectedHash: "0x456{}[]",
			actualHash:   "0x789<>?",
			wantTitle:    "[test!@#] Split hash has changed",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test!@#
Group: test$%^
Split Address: 0x123&*()
Expected Hash: 0x456{}[]
Actual Hash: 0x789<>?`,
		},
		{
			name:         "empty hashes",
			timestamp:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:      "test_monitor",
			group:        "test_group",
			splitAddress: "",
			expectedHash: "",
			actualHash:   "",
			wantTitle:    "[test_monitor] Split hash has changed",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Split Address: 
Expected Hash: 
Actual Hash: `,
		},
		{
			name:         "very long strings",
			timestamp:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:      strings.Repeat("m", 100),
			group:        strings.Repeat("g", 100),
			splitAddress: "0x" + strings.Repeat("1", 64),
			expectedHash: "0x" + strings.Repeat("2", 64),
			actualHash:   "0x" + strings.Repeat("3", 64),
			wantTitle:    "[" + strings.Repeat("m", 100) + "] Split hash has changed",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: ` + strings.Repeat("m", 100) + `
Group: ` + strings.Repeat("g", 100) + `
Split Address: 0x` + strings.Repeat("1", 64) + `
Expected Hash: 0x` + strings.Repeat("2", 64) + `
Actual Hash: 0x` + strings.Repeat("3", 64),
		},
		{
			name:         "unicode characters",
			timestamp:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:      "测试监控器",
			group:        "测试组",
			splitAddress: "0x测试地址",
			expectedHash: "0x预期哈希",
			actualHash:   "0x实际哈希",
			wantTitle:    "[测试监控器] Split hash has changed",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: 测试监控器
Group: 测试组
Split Address: 0x测试地址
Expected Hash: 0x预期哈希
Actual Hash: 0x实际哈希`,
		},
		{
			name:         "edge timestamp",
			timestamp:    time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC),
			monitor:      "test_monitor",
			group:        "test_group",
			splitAddress: "0x123",
			expectedHash: "0x456",
			actualHash:   "0x789",
			wantTitle:    "[test_monitor] Split hash has changed",
			wantDesc: `
Timestamp: 9999-12-31 23:59:59 UTC
Monitor: test_monitor
Group: test_group
Split Address: 0x123
Expected Hash: 0x456
Actual Hash: 0x789`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := split.NewHash(
				tt.timestamp,
				tt.monitor,
				tt.group,
				tt.splitAddress,
				tt.expectedHash,
				tt.actualHash,
			)

			// Verify it implements Event interface
			var _ event.Event = evt

			// Test type constant
			assert.Equal(t, split.HashType, evt.GetType())

			// Test getters
			assert.Equal(t, tt.monitor, evt.GetMonitor())
			assert.Equal(t, tt.group, evt.GetGroup())
			assert.Equal(t, tt.wantTitle, evt.GetTitle(true, true))
			assert.Equal(t, tt.wantDesc, evt.GetDescriptionText(true, true))

			// Test fields
			assert.Equal(t, tt.timestamp, evt.Timestamp)
			assert.Equal(t, tt.splitAddress, evt.SplitAddress)
			assert.Equal(t, tt.expectedHash, evt.ExpectedHash)
			assert.Equal(t, tt.actualHash, evt.ActualHash)
		})
	}
}
