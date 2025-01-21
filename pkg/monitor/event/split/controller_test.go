package split_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/ethpandaops/splitoor/pkg/monitor/event/split"
)

func TestController(t *testing.T) {
	tests := []struct {
		name               string
		timestamp          time.Time
		monitor            string
		group              string
		splitAddress       string
		expectedController string
		actualController   string
		wantTitle          string
		wantDesc           string
	}{
		{
			name:               "basic event",
			timestamp:          time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:            "test_monitor",
			group:              "test_group",
			splitAddress:       "0x123",
			expectedController: "0x456",
			actualController:   "0x789",
			wantTitle:          "[test_monitor] test_group split controller has changed",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Split Address: 0x123
Expected Controller address: 0x456
Actual Controller address: 0x789`,
		},
		{
			name:               "same controller",
			timestamp:          time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:            "test_monitor",
			group:              "test_group",
			splitAddress:       "0x123",
			expectedController: "0x456",
			actualController:   "0x456",
			wantTitle:          "[test_monitor] test_group split controller has changed",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Split Address: 0x123
Expected Controller address: 0x456
Actual Controller address: 0x456`,
		},
		{
			name:               "special characters",
			timestamp:          time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:            "test!@#",
			group:              "test$%^",
			splitAddress:       "0x123&*()",
			expectedController: "0x456{}[]",
			actualController:   "0x789<>?",
			wantTitle:          "[test!@#] test$%^ split controller has changed",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test!@#
Group: test$%^
Split Address: 0x123&*()
Expected Controller address: 0x456{}[]
Actual Controller address: 0x789<>?`,
		},
		{
			name:               "empty addresses",
			timestamp:          time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:            "test_monitor",
			group:              "test_group",
			splitAddress:       "",
			expectedController: "",
			actualController:   "",
			wantTitle:          "[test_monitor] test_group split controller has changed",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: test_monitor
Group: test_group
Split Address: 
Expected Controller address: 
Actual Controller address: `,
		},
		{
			name:               "very long strings",
			timestamp:          time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:            strings.Repeat("m", 100),
			group:              strings.Repeat("g", 100),
			splitAddress:       "0x" + strings.Repeat("1", 64),
			expectedController: "0x" + strings.Repeat("2", 64),
			actualController:   "0x" + strings.Repeat("3", 64),
			wantTitle:          "[" + strings.Repeat("m", 100) + "] " + strings.Repeat("g", 100) + " split controller has changed",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: ` + strings.Repeat("m", 100) + `
Group: ` + strings.Repeat("g", 100) + `
Split Address: 0x` + strings.Repeat("1", 64) + `
Expected Controller address: 0x` + strings.Repeat("2", 64) + `
Actual Controller address: 0x` + strings.Repeat("3", 64),
		},
		{
			name:               "unicode characters",
			timestamp:          time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			monitor:            "测试监控器",
			group:              "测试组",
			splitAddress:       "0x测试地址",
			expectedController: "0x预期控制器",
			actualController:   "0x实际控制器",
			wantTitle:          "[测试监控器] 测试组 split controller has changed",
			wantDesc: `
Timestamp: 2024-01-01 12:00:00 UTC
Monitor: 测试监控器
Group: 测试组
Split Address: 0x测试地址
Expected Controller address: 0x预期控制器
Actual Controller address: 0x实际控制器`,
		},
		{
			name:               "edge timestamp",
			timestamp:          time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC),
			monitor:            "test_monitor",
			group:              "test_group",
			splitAddress:       "0x123",
			expectedController: "0x456",
			actualController:   "0x789",
			wantTitle:          "[test_monitor] test_group split controller has changed",
			wantDesc: `
Timestamp: 9999-12-31 23:59:59 UTC
Monitor: test_monitor
Group: test_group
Split Address: 0x123
Expected Controller address: 0x456
Actual Controller address: 0x789`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := split.NewController(
				tt.timestamp,
				tt.monitor,
				tt.group,
				tt.splitAddress,
				tt.expectedController,
				tt.actualController,
			)

			// Verify it implements Event interface
			var _ event.Event = evt

			// Test type constant
			assert.Equal(t, split.ControllerType, evt.GetType())

			// Test getters
			assert.Equal(t, tt.monitor, evt.GetMonitor())
			assert.Equal(t, tt.group, evt.GetGroup())
			assert.Equal(t, tt.wantTitle, evt.GetTitle())
			assert.Equal(t, tt.wantDesc, evt.GetDescription())

			// Test fields
			assert.Equal(t, tt.timestamp, evt.Timestamp)
			assert.Equal(t, tt.splitAddress, evt.SplitAddress)
			assert.Equal(t, tt.expectedController, evt.ExpectedController)
			assert.Equal(t, tt.actualController, evt.ActualController)
		})
	}
}
