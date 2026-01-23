package source

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	testMonitorName = "test-monitor"
	testSourceName  = "test-source"
)

func TestNewSourceInvalidType(t *testing.T) {
	// Setup logger
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	ctx := context.Background()
	sourceType := SourceTypeUnknown // Invalid type
	includeMonitorName := true
	includeGroupName := true

	// Should return error for unknown source type
	_, err := NewSource(ctx, log, testMonitorName, testSourceName, nil, sourceType, includeMonitorName, includeGroupName, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source type is required")
}

func TestNewSourceUnsupportedType(t *testing.T) {
	// Setup logger
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	ctx := context.Background()
	sourceType := SourceType("unsupported") // Valid but unsupported type
	includeMonitorName := true
	includeGroupName := true

	// Should return error for unsupported source type
	_, err := NewSource(ctx, log, testMonitorName, testSourceName, nil, sourceType, includeMonitorName, includeGroupName, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source type is not supported")
}

func TestNewSourceConfigUnmarshalError(t *testing.T) {
	// Setup logger
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress log output during tests

	ctx := context.Background()
	sourceType := SourceTypeDiscord // Valid type
	includeMonitorName := true
	includeGroupName := true

	// Create a RawMessage that will return an error
	mockRawMsg := &RawMessage{
		unmarshal: func(any) error {
			return errors.New("mock unmarshal error")
		},
	}

	// Should return error when config can't be unmarshalled
	_, err := NewSource(ctx, log, testMonitorName, testSourceName, nil, sourceType, includeMonitorName, includeGroupName, mockRawMsg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock unmarshal error")
}

// Test the Config validation.
func TestConfigValidate(t *testing.T) {
	// Test with unknown source type
	config := &Config{
		SourceType: SourceTypeUnknown,
		Name:       "test",
	}
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notifier source type is required")

	// Test with valid source type
	config = &Config{
		SourceType: SourceTypeDiscord,
		Name:       "test",
	}
	err = config.Validate()
	assert.NoError(t, err)
}

// Test RawMessage.
func TestRawMessageUnmarshal(t *testing.T) {
	// Create a raw message with a mock unmarshal function
	var called bool

	rawMsg := &RawMessage{
		unmarshal: func(v any) error {
			called = true

			return nil
		},
	}

	// Call Unmarshal
	err := rawMsg.Unmarshal(nil)
	assert.NoError(t, err)
	assert.True(t, called, "Unmarshal function should have been called")

	// Test with a failing unmarshal function
	expectedErr := errors.New("unmarshal error")
	rawMsg = &RawMessage{
		unmarshal: func(v any) error {
			return expectedErr
		},
	}

	err = rawMsg.Unmarshal(nil)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestRawMessageUnmarshalYAML(t *testing.T) {
	// Create a raw message
	rawMsg := &RawMessage{}

	// Create a mock unmarshal function
	mockUnmarshal := func(v any) error {
		return nil
	}

	// Call UnmarshalYAML
	err := rawMsg.UnmarshalYAML(mockUnmarshal)
	assert.NoError(t, err)

	// Verify that the unmarshal function was set
	assert.NotNil(t, rawMsg.unmarshal)

	// Call Unmarshal to verify it uses the set function
	err = rawMsg.Unmarshal(nil)
	assert.NoError(t, err)
}

func TestRawMessageNilUnmarshal(t *testing.T) {
	// Create a raw message with nil unmarshal function
	rawMsg := &RawMessage{}

	// Call Unmarshal should panic since unmarshal is nil
	// We recover from the panic and verify it occurred
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Expected panic from nil unmarshal function")
	}()

	_ = rawMsg.Unmarshal(nil)
}
