package discord_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	disc "github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/discord"
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

func TestNewDiscord(t *testing.T) {
	tests := []struct {
		name        string
		monitor     string
		config      *disc.Config
		expectError bool
	}{
		{
			name:    "valid config",
			monitor: "test_monitor",
			config: &disc.Config{
				Webhook: "https://discord.com/api/webhooks/test",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logrus.New()
			entry := log.WithField("test", "test")
			discord, err := disc.NewDiscord(context.Background(), entry, tt.monitor, tt.name, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, discord)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, discord)
				assert.Equal(t, tt.name, discord.GetName())
				assert.Equal(t, tt.config.Webhook, discord.GetConfig().Webhook)
			}
		})
	}
}

func TestDiscordPublish(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse int
		expectError    bool
	}{
		{
			name:           "successful publish",
			serverResponse: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server error",
			serverResponse: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "unauthorized",
			serverResponse: http.StatusUnauthorized,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				w.WriteHeader(tt.serverResponse)
			}))
			defer server.Close()

			log := logrus.New()
			entry := log.WithField("test", "test")
			discord, err := disc.NewDiscord(context.Background(), entry, "test", "test_source", &disc.Config{
				Webhook: server.URL,
			})
			assert.NoError(t, err)

			mockEvent := new(MockEvent)
			mockEvent.On("GetGroup").Return("test_group").Times(2)
			mockEvent.On("GetTitle").Return("Test Title").Once()
			mockEvent.On("GetDescription").Return("Test Description").Once()

			err = discord.Publish(context.Background(), mockEvent)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockEvent.AssertExpectations(t)
		})
	}
}

func TestDiscordStartStop(t *testing.T) {
	log := logrus.New()
	entry := log.WithField("test", "test")
	discord, err := disc.NewDiscord(context.Background(), entry, "test", "test_source", &disc.Config{
		Webhook: "https://discord.com/api/webhooks/test",
	})
	assert.NoError(t, err)

	err = discord.Start(context.Background())
	assert.NoError(t, err)

	err = discord.Stop(context.Background())
	assert.NoError(t, err)
}

func TestDiscordGetTypeAndName(t *testing.T) {
	log := logrus.New()
	entry := log.WithField("test", "test")
	discord, err := disc.NewDiscord(context.Background(), entry, "test", "test_source", &disc.Config{
		Webhook: "https://discord.com/api/webhooks/test",
	})
	assert.NoError(t, err)

	assert.Equal(t, disc.SourceType, discord.GetType())
	assert.Equal(t, "test_source", discord.GetName())
}
