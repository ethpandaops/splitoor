package ses_test

import (
	"context"
	"testing"

	s "github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/ses"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockEvent) GetTitle(includeMonitorName, includeGroupName bool) string {
	args := m.Called()

	return args.String(0)
}

func (m *MockEvent) GetDescriptionText(includeMonitorName, includeGroupName bool) string {
	args := m.Called()

	return args.String(0)
}

func (m *MockEvent) GetDescriptionMarkdown(includeMonitorName, includeGroupName bool) string {
	args := m.Called()

	return args.String(0)
}

func TestNewSES(t *testing.T) {
	tests := []struct {
		name        string
		monitor     string
		config      *s.Config
		expectError bool
	}{
		{
			name:    "valid config",
			monitor: "test_monitor",
			config: &s.Config{
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
			ses, err := s.NewSES(context.Background(), entry, tt.monitor, tt.name, nil, true, true, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, ses)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ses)
				assert.Equal(t, tt.name, ses.GetName())
				assert.Equal(t, tt.config, ses.GetConfig())
			}
		})
	}
}

func TestSESStartStop(t *testing.T) {
	log := logrus.New()
	entry := log.WithField("test", "test")
	ses, err := s.NewSES(context.Background(), entry, "test", "test_source", nil, true, true, &s.Config{
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
	ses, err := s.NewSES(context.Background(), entry, "test", "test_source", nil, true, true, &s.Config{
		From: "test@example.com",
		To:   []string{"recipient@example.com"},
	})
	assert.NoError(t, err)

	assert.Equal(t, s.SourceType, ses.GetType())
	assert.Equal(t, "test_source", ses.GetName())
}
