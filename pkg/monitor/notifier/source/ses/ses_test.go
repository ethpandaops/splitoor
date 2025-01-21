package ses

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const SourceType = "ses"

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

func (m *MockEvent) GetGroup() string {
	args := m.Called()

	return args.String(0)
}

func (m *MockEvent) GetTitle() string {
	args := m.Called()

	return args.String(0)
}

func (m *MockEvent) GetDescription() string {
	args := m.Called()

	return args.String(0)
}

func TestNewSES(t *testing.T) {
	tests := []struct {
		name        string
		monitor     string
		config      *Config
		expectError bool
	}{
		{
			name:    "valid config",
			monitor: "test_monitor",
			config: &Config{
				From: "test@example.com",
				To:   []string{"recipient@example.com"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logrus.New()
			entry := log.WithField("test", "test")
			ses, err := NewSES(context.Background(), entry, tt.monitor, tt.name, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, ses)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ses)
				assert.Equal(t, tt.name, ses.name)
				assert.NotNil(t, ses.config)
				assert.Equal(t, tt.config, ses.config)
			}
		})
	}
}

func TestSESStartStop(t *testing.T) {
	log := logrus.New()
	entry := log.WithField("test", "test")
	ses, err := NewSES(context.Background(), entry, "test", "test_source", &Config{
		From: "test@example.com",
		To:   []string{"recipient@example.com"},
	})
	assert.NoError(t, err)

	err = ses.Start(context.Background())
	assert.NoError(t, err)

	err = ses.Stop(context.Background())
	assert.NoError(t, err)
}

func TestSESGetTypeAndName(t *testing.T) {
	log := logrus.New()
	entry := log.WithField("test", "test")
	ses, err := NewSES(context.Background(), entry, "test", "test_source", &Config{
		From: "test@example.com",
		To:   []string{"recipient@example.com"},
	})
	assert.NoError(t, err)

	assert.Equal(t, SourceType, ses.GetType())
	assert.Equal(t, ses.name, ses.GetName())
}
