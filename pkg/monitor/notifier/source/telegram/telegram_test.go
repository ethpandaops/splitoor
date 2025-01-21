package telegram

import (
	"context"
	"testing"

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
	*bot.Bot
}

func (m *MockBot) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	args := m.Called(ctx, params)

	return args.Get(0).(*models.Message), args.Error(1)
}

func TestNewTelegram(t *testing.T) {
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
				BotToken: "test-token",
				ChatID:   "123456789",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logrus.New()
			entry := log.WithField("test", "test")
			mockBot := &MockBot{}
			mockBot.On("SendMessage", mock.Anything, mock.Anything).Return(&models.Message{}, nil)

			telegram := &Telegram{
				log:     entry.WithField("source", "telegram"),
				config:  tt.config,
				monitor: tt.monitor,
				name:    tt.name,
				client:  mockBot.Bot,
				metrics: GetMetricsInstance("splitoor_notifier_telegram", tt.monitor),
			}

			assert.NotNil(t, telegram)
			assert.Equal(t, tt.name, telegram.name)
			assert.NotNil(t, telegram.config)
			assert.Equal(t, tt.config, telegram.config)
		})
	}
}

func TestTelegramStartStop(t *testing.T) {
	log := logrus.New()
	entry := log.WithField("test", "test")
	mockBot := &MockBot{}
	mockBot.On("SendMessage", mock.Anything, mock.Anything).Return(&models.Message{}, nil)

	telegram := &Telegram{
		log:     entry.WithField("source", "telegram"),
		config:  &Config{BotToken: "test-token", ChatID: "123456789"},
		monitor: "test",
		name:    "test_source",
		client:  mockBot.Bot,
		metrics: GetMetricsInstance("splitoor_notifier_telegram", "test"),
	}

	err := telegram.Start(context.Background())
	assert.NoError(t, err)

	err = telegram.Stop(context.Background())
	assert.NoError(t, err)
}

func TestTelegramGetTypeAndName(t *testing.T) {
	log := logrus.New()
	entry := log.WithField("test", "test")
	mockBot := &MockBot{}
	mockBot.On("SendMessage", mock.Anything, mock.Anything).Return(&models.Message{}, nil)

	telegram := &Telegram{
		log:     entry.WithField("source", "telegram"),
		config:  &Config{BotToken: "test-token", ChatID: "123456789"},
		monitor: "test",
		name:    "test_source",
		client:  mockBot.Bot,
		metrics: GetMetricsInstance("splitoor_notifier_telegram", "test"),
	}

	assert.Equal(t, SourceType, telegram.GetType())
	assert.Equal(t, telegram.name, telegram.GetName())
}
