package smtp_test

import (
	"context"
	"testing"

	email "github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/smtp"
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

func (m *MockEvent) GetTitle() string {
	args := m.Called()

	return args.String(0)
}

func (m *MockEvent) GetDescription() string {
	args := m.Called()

	return args.String(0)
}

func TestNewSMTP(t *testing.T) {
	tests := []struct {
		name        string
		monitor     string
		config      *email.Config
		expectError bool
	}{
		{
			name:    "valid config",
			monitor: "test_monitor",
			config: &email.Config{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
				From:     "test@example.com",
				To:       []string{"recipient@example.com"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logrus.New()
			entry := log.WithField("test", "test")
			smtp, err := email.NewSMTP(context.Background(), entry, tt.monitor, tt.name, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, smtp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, smtp)
				assert.Equal(t, tt.name, smtp.GetName())
				assert.Equal(t, tt.config, smtp.GetConfig())
			}
		})
	}
}

func TestSMTPStartStop(t *testing.T) {
	log := logrus.New()
	entry := log.WithField("test", "test")
	smtp, err := email.NewSMTP(context.Background(), entry, "test", "test_source", &email.Config{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		To:       []string{"recipient@example.com"},
	})
	assert.NoError(t, err)

	err = smtp.Start(context.Background())
	assert.NoError(t, err)

	err = smtp.Stop(context.Background())
	assert.NoError(t, err)
}

func TestSMTPGetTypeAndName(t *testing.T) {
	log := logrus.New()
	entry := log.WithField("test", "test")
	smtp, err := email.NewSMTP(context.Background(), entry, "test", "test_source", &email.Config{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		To:       []string{"recipient@example.com"},
	})
	assert.NoError(t, err)

	assert.Equal(t, email.SourceType, smtp.GetType())
	assert.Equal(t, "test_source", smtp.GetName())
}
