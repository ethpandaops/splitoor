package telegram_test

import (
	"context"
	"testing"

	tel "github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/telegram"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const SourceType = "telegram"

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

type MockBot struct {
	mock.Mock
}

func (m *MockBot) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	args := m.Called(ctx, params)

	return args.Get(0).(*models.Message), args.Error(1)
}

func setupMockBot() *MockBot {
	mockBot := &MockBot{}
	mockBot.On("SendMessage", mock.Anything, mock.Anything).Return(&models.Message{}, nil)

	return mockBot
}

func TestNewTelegram(t *testing.T) {
	tests := []struct {
		name        string
		monitor     string
		config      *tel.Config
		expectError bool
	}{
		{
			name:    "valid config",
			monitor: "test_monitor",
			config: &tel.Config{
				BotToken: "test-token",
				ChatID:   "123456789",
			},
			expectError: false,
		},
		{
			name:    "invalid config - empty token",
			monitor: "test_monitor",
			config: &tel.Config{
				BotToken: "",
				ChatID:   "123456789",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logrus.New()
			entry := log.WithField("test", "test")

			if tt.expectError {
				telegram, err := tel.NewTelegram(context.Background(), entry.WithField("source", "telegram"), tt.monitor, tt.name, tt.config)
				assert.Error(t, err)
				assert.Nil(t, telegram)

				return
			}

			mockBot := setupMockBot()
			telegram, err := tel.NewTelegramWithClient(entry.WithField("source", "telegram"), tt.monitor, tt.name, tt.config, mockBot)
			assert.NoError(t, err)
			assert.NotNil(t, telegram)
			assert.Equal(t, tt.name, telegram.GetName())
			assert.NotNil(t, telegram.GetConfig())
			assert.Equal(t, tt.config, telegram.GetConfig())
		})
	}
}

func TestTelegramPublish(t *testing.T) {
	log := logrus.New()
	entry := log.WithField("test", "test")
	mockBot := &MockBot{}
	mockEvent := &MockEvent{}

	// Setup mock event
	mockEvent.On("GetGroup").Return("test_group")
	mockEvent.On("GetTitle").Return("Test Title")
	mockEvent.On("GetDescription").Return("Test Description")

	// Setup mock bot expectations
	mockBot.On("SendMessage", mock.Anything, &bot.SendMessageParams{
		ChatID:    int64(123456789),
		Text:      "<b>Test Title</b>\n\nTest Description",
		ParseMode: "HTML",
	}).Return(&models.Message{}, nil)

	telegram, err := tel.NewTelegramWithClient(entry.WithField("source", "telegram"), "test", "test_source", &tel.Config{
		BotToken: "test-token",
		ChatID:   "123456789",
	}, mockBot)
	assert.NoError(t, err)

	// Test Publish
	err = telegram.Publish(context.Background(), mockEvent)
	assert.NoError(t, err)

	mockBot.AssertExpectations(t)
	mockEvent.AssertExpectations(t)
}

func TestTelegramStartStop(t *testing.T) {
	log := logrus.New()
	entry := log.WithField("test", "test")
	mockBot := setupMockBot()

	telegram, err := tel.NewTelegramWithClient(entry.WithField("source", "telegram"), "test", "test_source", &tel.Config{BotToken: "test-token", ChatID: "123456789"}, mockBot)
	assert.NoError(t, err)

	err = telegram.Start(context.Background())
	assert.NoError(t, err)

	err = telegram.Stop(context.Background())
	assert.NoError(t, err)
}

func TestTelegramGetTypeAndName(t *testing.T) {
	log := logrus.New()
	entry := log.WithField("test", "test")
	mockBot := setupMockBot()

	telegram, err := tel.NewTelegramWithClient(entry.WithField("source", "telegram"), "test", "test_source", &tel.Config{BotToken: "test-token", ChatID: "123456789"}, mockBot)
	assert.NoError(t, err)

	assert.Equal(t, SourceType, telegram.GetType())
	assert.Equal(t, "test_source", telegram.GetName())
}
