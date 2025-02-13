package event_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
)

// MockEvent implements Event interface for testing
type MockEvent struct {
	monitor     string
	eventType   string
	title       string
	description string
	group       string
}

func NewMockEvent(monitor, eventType, title, description, group string) *MockEvent {
	return &MockEvent{
		monitor:     monitor,
		eventType:   eventType,
		title:       title,
		description: description,
		group:       group,
	}
}

func (m *MockEvent) GetMonitor() string {
	return m.monitor
}

func (m *MockEvent) GetType() string {
	return m.eventType
}

func (m *MockEvent) GetTitle(includeMonitorName, includeGroupName bool) string {
	return m.title
}

func (m *MockEvent) GetDescriptionText(includeMonitorName, includeGroupName bool) string {
	return m.description
}

func (m *MockEvent) GetDescriptionMarkdown(includeMonitorName, includeGroupName bool) string {
	return m.description
}

func (m *MockEvent) GetDescriptionHTML(includeMonitorName, includeGroupName bool) string {
	return m.description
}

func (m *MockEvent) GetGroup() string {
	return m.group
}

func TestEventInterface(t *testing.T) {
	tests := []struct {
		name        string
		monitor     string
		eventType   string
		title       string
		description string
		group       string
	}{
		{
			name:        "basic event",
			monitor:     "test_monitor",
			eventType:   "test_type",
			title:       "Test Title",
			description: "Test Description",
			group:       "test_group",
		},
		{
			name:        "empty fields",
			monitor:     "",
			eventType:   "",
			title:       "",
			description: "",
			group:       "",
		},
		{
			name:        "special characters",
			monitor:     "test-monitor_123",
			eventType:   "test.type@123",
			title:       "Test Title !@#$%^&*()",
			description: "Test Description\nWith\nNewlines",
			group:       "test/group/123",
		},
		{
			name:        "very long fields",
			monitor:     strings.Repeat("a", 1000),
			eventType:   strings.Repeat("b", 1000),
			title:       strings.Repeat("c", 1000),
			description: strings.Repeat("d\n", 1000),
			group:       strings.Repeat("e", 1000),
		},
		{
			name:        "unicode characters",
			monitor:     "测试监控器",
			eventType:   "测试类型",
			title:       "测试标题",
			description: "测试描述\n换行",
			group:       "测试组",
		},
		{
			name:        "json special characters",
			monitor:     `{"key": "value"}`,
			eventType:   `{"type": "test"}`,
			title:       `{"title": "test"}`,
			description: `{"desc": "test"}`,
			group:       `{"group": "test"}`,
		},
		{
			name:        "html characters",
			monitor:     "<monitor>test</monitor>",
			eventType:   "<type>test</type>",
			title:       "<title>test</title>",
			description: "<desc>test</desc>",
			group:       "<group>test</group>",
		},
		{
			name:        "sql injection characters",
			monitor:     "'; DROP TABLE events; --",
			eventType:   "'; DELETE FROM events; --",
			title:       "'; TRUNCATE events; --",
			description: "'; SELECT * FROM events; --",
			group:       "'; UPDATE events SET type='hacked'; --",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEvent := NewMockEvent(tt.monitor, tt.eventType, tt.title, tt.description, tt.group)

			// Verify the mock implements Event interface
			var _ event.Event = mockEvent

			// Test all getters
			assert.Equal(t, tt.monitor, mockEvent.GetMonitor())
			assert.Equal(t, tt.eventType, mockEvent.GetType())
			assert.Equal(t, tt.title, mockEvent.GetTitle(true, true))
			assert.Equal(t, tt.description, mockEvent.GetDescriptionText(true, true))
			assert.Equal(t, tt.description, mockEvent.GetDescriptionMarkdown(true, true))
			assert.Equal(t, tt.description, mockEvent.GetDescriptionHTML(true, true))
			assert.Equal(t, tt.group, mockEvent.GetGroup())

			// Test string representations
			require.NotPanics(t, func() {
				_ = fmt.Sprintf("%+v", mockEvent)
				_ = fmt.Sprintf("%v", mockEvent)
				_ = fmt.Sprintf("%s", mockEvent)
			})
		})
	}
}

// TestEventComparison tests comparing different events
func TestEventComparison(t *testing.T) {
	baseEvent := NewMockEvent("monitor", "type", "title", "desc", "group")

	tests := []struct {
		name     string
		event1   event.Event
		event2   event.Event
		shouldEq bool
	}{
		{
			name:     "identical events",
			event1:   baseEvent,
			event2:   NewMockEvent("monitor", "type", "title", "desc", "group"),
			shouldEq: true,
		},
		{
			name:     "different monitor",
			event1:   baseEvent,
			event2:   NewMockEvent("other", "type", "title", "desc", "group"),
			shouldEq: false,
		},
		{
			name:     "different type",
			event1:   baseEvent,
			event2:   NewMockEvent("monitor", "other", "title", "desc", "group"),
			shouldEq: false,
		},
		{
			name:     "different title",
			event1:   baseEvent,
			event2:   NewMockEvent("monitor", "type", "other", "desc", "group"),
			shouldEq: false,
		},
		{
			name:     "different description",
			event1:   baseEvent,
			event2:   NewMockEvent("monitor", "type", "title", "other", "group"),
			shouldEq: false,
		},
		{
			name:     "different group",
			event1:   baseEvent,
			event2:   NewMockEvent("monitor", "type", "title", "desc", "other"),
			shouldEq: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test equality
			assert.Equal(t, tt.shouldEq,
				tt.event1.GetMonitor() == tt.event2.GetMonitor() &&
					tt.event1.GetType() == tt.event2.GetType() &&
					tt.event1.GetTitle(true, true) == tt.event2.GetTitle(true, true) &&
					tt.event1.GetDescriptionText(true, true) == tt.event2.GetDescriptionText(true, true) &&
					tt.event1.GetDescriptionMarkdown(true, true) == tt.event2.GetDescriptionMarkdown(true, true) &&
					tt.event1.GetDescriptionHTML(true, true) == tt.event2.GetDescriptionHTML(true, true) &&
					tt.event1.GetGroup() == tt.event2.GetGroup())
		})
	}
}

// TestEventValidation tests validation of event fields
func TestEventValidation(t *testing.T) {
	tests := []struct {
		name       string
		event      event.Event
		shouldBeOK bool
	}{
		{
			name: "valid event",
			event: NewMockEvent(
				"monitor",
				"type",
				"title",
				"description",
				"group",
			),
			shouldBeOK: true,
		},
		{
			name: "missing monitor",
			event: NewMockEvent(
				"",
				"type",
				"title",
				"description",
				"group",
			),
			shouldBeOK: false,
		},
		{
			name: "missing type",
			event: NewMockEvent(
				"monitor",
				"",
				"title",
				"description",
				"group",
			),
			shouldBeOK: false,
		},
		{
			name: "missing title",
			event: NewMockEvent(
				"monitor",
				"type",
				"",
				"description",
				"group",
			),
			shouldBeOK: false,
		},
		{
			name: "missing description",
			event: NewMockEvent(
				"monitor",
				"type",
				"title",
				"",
				"group",
			),
			shouldBeOK: false,
		},
		{
			name: "missing group",
			event: NewMockEvent(
				"monitor",
				"type",
				"title",
				"description",
				"",
			),
			shouldBeOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if required fields are present
			hasAllFields := tt.event.GetMonitor() != "" &&
				tt.event.GetType() != "" &&
				tt.event.GetTitle(true, true) != "" &&
				tt.event.GetDescriptionText(true, true) != "" &&
				tt.event.GetDescriptionMarkdown(true, true) != "" &&
				tt.event.GetDescriptionHTML(true, true) != "" &&
				tt.event.GetGroup() != ""

			assert.Equal(t, tt.shouldBeOK, hasAllFields)
		})
	}
}
