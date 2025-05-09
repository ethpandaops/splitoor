package notifier

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEvent implements the event.Event interface for testing
type MockEvent struct {
	mock.Mock
}

func (m *MockEvent) GetMonitor() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEvent) GetType() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEvent) GetTitle(includeMonitor, includeGroup bool) string {
	args := m.Called(includeMonitor, includeGroup)
	return args.String(0)
}

func (m *MockEvent) GetDescriptionText(includeMonitor, includeGroup bool) string {
	args := m.Called(includeMonitor, includeGroup)
	return args.String(0)
}

func (m *MockEvent) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
	args := m.Called(includeMonitor, includeGroup)
	return args.String(0)
}

func (m *MockEvent) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
	args := m.Called(includeMonitor, includeGroup)
	return args.String(0)
}

func (m *MockEvent) GetGroup() string {
	args := m.Called()
	return args.String(0)
}

// MockSource implements the source.Source interface for testing
type MockSource struct {
	mock.Mock
}

func (m *MockSource) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSource) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSource) GetType() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSource) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSource) Publish(ctx context.Context, e event.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func TestPublisherNilChecks(t *testing.T) {
	// Setup logger
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create publisher with no sources
	publisher := &Publisher{
		log:     log,
		sources: []SourceWithConfig{},
	}

	// Test nil event
	err := publisher.Publish(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot publish nil event")

	// Test nil context using PublishWithContext
	mockEvent := new(MockEvent)
	mockEvent.On("GetGroup").Return("test-group")
	err = publisher.PublishWithContext(nil, mockEvent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot publish with nil context")
}

func TestPublisherSourceGroupFiltering(t *testing.T) {
	// Setup logger
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create mock source
	mockSource := new(MockSource)
	mockSource.On("GetName").Return("test-source")
	mockSource.On("Publish", mock.Anything, mock.Anything).Return(nil)

	// Setup mock event
	mockEvent := new(MockEvent)
	mockEvent.On("GetGroup").Return("test-group")

	// Create publisher with a source that has a different group filter
	differentGroup := "different-group"
	publisher := &Publisher{
		log: log,
		sources: []SourceWithConfig{
			{
				source: mockSource,
				group:  &differentGroup,
			},
		},
	}

	// The event should be filtered out based on group
	err := publisher.Publish(mockEvent)
	assert.NoError(t, err)
	mockSource.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)

	// Create publisher with a source that has a matching group filter
	matchingGroup := "test-group"
	publisher = &Publisher{
		log: log,
		sources: []SourceWithConfig{
			{
				source: mockSource,
				group:  &matchingGroup,
			},
		},
	}

	// The event should be published
	err = publisher.Publish(mockEvent)
	assert.NoError(t, err)
	mockSource.AssertCalled(t, "Publish", mock.Anything, mockEvent)
}

func TestPublisherNilSource(t *testing.T) {
	// Setup logger
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Setup mock event
	mockEvent := new(MockEvent)
	mockEvent.On("GetGroup").Return("test-group")

	// Create publisher with a nil source
	publisher := &Publisher{
		log: log,
		sources: []SourceWithConfig{
			{
				source: nil,
				group:  nil,
			},
		},
	}

	// The nil source should be skipped without error
	err := publisher.Publish(mockEvent)
	assert.NoError(t, err)
}

func TestPublisherContextCancellation(t *testing.T) {
	// Setup logger
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create mock source that takes time to publish
	mockSource := new(MockSource)
	mockSource.On("GetName").Return("slow-source")
	mockSource.On("Publish", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(100 * time.Millisecond)
	}).Return(nil)

	// Setup mock event
	mockEvent := new(MockEvent)
	mockEvent.On("GetGroup").Return("test-group")

	// Create publisher with the slow mock source
	publisher := &Publisher{
		log: log,
		sources: []SourceWithConfig{
			{
				source: mockSource,
				group:  nil,
			},
		},
	}

	// Create a context that will be canceled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The publish operation should fail due to context cancellation
	err := publisher.PublishWithContext(ctx, mockEvent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
}

func TestPublisherSourceError(t *testing.T) {
	// Setup logger
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create mock source that returns an error
	mockSource := new(MockSource)
	mockSource.On("GetName").Return("error-source")
	expectedErr := fmt.Errorf("publish error")
	mockSource.On("Publish", mock.Anything, mock.Anything).Return(expectedErr)

	// Setup mock event
	mockEvent := new(MockEvent)
	mockEvent.On("GetGroup").Return("test-group")

	// Create publisher with the error-returning source
	publisher := &Publisher{
		log: log,
		sources: []SourceWithConfig{
			{
				source: mockSource,
				group:  nil,
			},
		},
	}

	// The publish operation should return the source error
	err := publisher.Publish(mockEvent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error publishing to source")
}

func TestPublisherStartStop(t *testing.T) {
	// Setup logger
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create mock sources
	mockSourceA := new(MockSource)
	mockSourceA.On("Start", mock.Anything).Return(nil)
	mockSourceA.On("Stop", mock.Anything).Return(nil)

	mockSourceB := new(MockSource)
	mockSourceB.On("Start", mock.Anything).Return(nil)
	mockSourceB.On("Stop", mock.Anything).Return(nil)

	// Create publisher with the mock sources and a nil source
	publisher := &Publisher{
		log: log,
		sources: []SourceWithConfig{
			{
				source: mockSourceA,
				group:  nil,
			},
			{
				source: nil, // Should be skipped
				group:  nil,
			},
			{
				source: mockSourceB,
				group:  nil,
			},
		},
	}

	// Test Start
	ctx := context.Background()
	err := publisher.Start(ctx)
	assert.NoError(t, err)
	mockSourceA.AssertCalled(t, "Start", ctx)
	mockSourceB.AssertCalled(t, "Start", ctx)

	// Test Stop
	err = publisher.Stop(ctx)
	assert.NoError(t, err)
	mockSourceA.AssertCalled(t, "Stop", ctx)
	mockSourceB.AssertCalled(t, "Stop", ctx)
}

func TestPublisherStartError(t *testing.T) {
	// Setup logger
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create mock source that returns an error on Start
	mockSource := new(MockSource)
	expectedErr := fmt.Errorf("start error")
	mockSource.On("Start", mock.Anything).Return(expectedErr)

	// Create publisher with the error-returning source
	publisher := &Publisher{
		log: log,
		sources: []SourceWithConfig{
			{
				source: mockSource,
				group:  nil,
			},
		},
	}

	// The Start operation should return the source error
	ctx := context.Background()
	err := publisher.Start(ctx)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestPublisherStopError(t *testing.T) {
	// Setup logger
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	// Create mock source that returns an error on Stop
	mockSource := new(MockSource)
	mockSource.On("Start", mock.Anything).Return(nil)
	expectedErr := fmt.Errorf("stop error")
	mockSource.On("Stop", mock.Anything).Return(expectedErr)

	// Create publisher with the error-returning source
	publisher := &Publisher{
		log: log,
		sources: []SourceWithConfig{
			{
				source: mockSource,
				group:  nil,
			},
		},
	}

	// Start should succeed
	ctx := context.Background()
	err := publisher.Start(ctx)
	assert.NoError(t, err)

	// The Stop operation should return the source error
	err = publisher.Stop(ctx)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
