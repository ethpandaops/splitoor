package telegram

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
)

const SourceType = "telegram"

type botClient interface {
	SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
}

type Telegram struct {
	log     logrus.FieldLogger
	config  *Config
	monitor string
	name    string
	client  botClient
	metrics *Metrics

	includeMonitorName bool
	includeGroupName   bool
	docs               *string
}

func NewTelegram(ctx context.Context, log logrus.FieldLogger, monitor, name string, docs *string, includeMonitorName, includeGroupName bool, config *Config) (*Telegram, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	client, err := bot.New(config.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return NewTelegramWithClient(log, monitor, name, docs, includeMonitorName, includeGroupName, config, client)
}

func NewTelegramWithClient(log logrus.FieldLogger, monitor, name string, docs *string, includeMonitorName, includeGroupName bool, config *Config, client botClient) (*Telegram, error) {
	return &Telegram{
		log:                log.WithField("source", "telegram"),
		config:             config,
		monitor:            monitor,
		name:               name,
		client:             client,
		metrics:            GetMetricsInstance("splitoor_notifier_telegram", monitor),
		includeMonitorName: includeMonitorName,
		includeGroupName:   includeGroupName,
		docs:               docs,
	}, nil
}

func (t *Telegram) WithClient(client botClient) *Telegram {
	t.client = client

	return t
}

func (t *Telegram) Start(ctx context.Context) error {
	return nil
}

func (t *Telegram) Stop(ctx context.Context) error {
	return nil
}

func (t *Telegram) GetType() string {
	return SourceType
}

func (t *Telegram) GetName() string {
	return t.name
}

func (t *Telegram) GetConfig() *Config {
	return t.config
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

	description := e.GetDescriptionMarkdown(t.includeMonitorName, t.includeGroupName)

	if t.docs != nil {
		docURL := strings.ReplaceAll(*t.docs, ":group", url.QueryEscape(e.GetGroup()))
		description = fmt.Sprintf("%s\n\n[**Go to docs**](%s)", description, docURL)
	}

	title := strings.ReplaceAll(e.GetTitle(t.includeMonitorName, t.includeGroupName), "-", "\\-")
	description = strings.ReplaceAll(strings.ReplaceAll(description, "**", "***"), "-", "\\-")
	text := fmt.Sprintf("🚨 ***%s***\n\n%s", title, description)

	isDisabled := true

	params := &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "MarkdownV2",
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: &isDisabled,
		},
	}

	if t.config.ThreadID != 0 {
		params.MessageThreadID = int(t.config.ThreadID)
	}

	_, err = t.client.SendMessage(ctx, params)

	if err != nil {
		errorType = "send_message"

		return fmt.Errorf("failed to send message: %w", err)
	}

	t.metrics.IncMessagesPublished(e.GetGroup(), t.name, t.GetType())

	return nil
}
