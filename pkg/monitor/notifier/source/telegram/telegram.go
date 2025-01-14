package telegram

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/go-telegram/bot"
	"github.com/sirupsen/logrus"
)

type Telegram struct {
	log     logrus.FieldLogger
	config  *Config
	monitor string
	name    string
	client  *bot.Bot
	metrics *Metrics
}

func NewTelegram(ctx context.Context, log logrus.FieldLogger, monitor, name string, config *Config) (*Telegram, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	client, err := bot.New(config.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return &Telegram{
		log:     log.WithField("source", "telegram"),
		config:  config,
		monitor: monitor,
		name:    name,
		client:  client,
		metrics: GetMetricsInstance("splitoor_notifier_telegram", monitor),
	}, nil
}

func (t *Telegram) Start(ctx context.Context) error {
	return nil
}

func (t *Telegram) Stop(ctx context.Context) error {
	return nil
}

func (t *Telegram) GetType() string {
	return "telegram"
}

func (t *Telegram) GetName() string {
	return t.name
}

func (t *Telegram) Publish(ctx context.Context, e event.Event) error {
	var errorType string
	defer func() {
		if errorType != "" {
			t.metrics.IncErrors(e.GetGroup(), t.name, t.GetType(), errorType)
		}
	}()

	chatID, err := strconv.ParseInt(t.config.ChatID, 10, 64)
	if err != nil {
		errorType = "invalid_chat_id"

		return fmt.Errorf("invalid chat ID: %w", err)
	}

	text := fmt.Sprintf("<b>%s</b>\n\n%s", e.GetTitle(), e.GetDescription())

	_, err = t.client.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	})

	if err != nil {
		errorType = "send_message"

		return fmt.Errorf("failed to send message: %w", err)
	}

	t.metrics.IncMessagesPublished(e.GetGroup(), t.name, t.GetType())

	return nil
}
