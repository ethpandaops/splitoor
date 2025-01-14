package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/sirupsen/logrus"
)

const SourceType = "discord"

type Discord struct {
	log        logrus.FieldLogger
	name       string
	sourceName string
	config     *Config
	metrics    *Metrics
}

func NewDiscord(ctx context.Context, log logrus.FieldLogger, monitor, sourceName string, config *Config) (*Discord, error) {
	return &Discord{
		log:        log.WithField("source", sourceName),
		sourceName: sourceName,
		config:     config,
		metrics:    GetMetricsInstance("splitoor_notifier_discord", monitor),
	}, nil
}

func (c *Discord) Start(ctx context.Context) error {
	return nil
}

func (c *Discord) Stop(ctx context.Context) error {
	return nil
}

func (c *Discord) GetType() string {
	return SourceType
}

func (c *Discord) GetName() string {
	return c.name
}

func (c *Discord) Publish(ctx context.Context, e event.Event) error {
	log := c.log.WithField("group", e.GetGroup())
	log.Info("Publishing message to Discord")

	var errorType string

	var statusCode string

	defer func() {
		if errorType != "" {
			c.metrics.IncErrors(e.GetGroup(), c.sourceName, c.GetType(), errorType, statusCode)
		} else {
			c.metrics.IncMessagesPublished(e.GetGroup(), c.sourceName, c.GetType())
		}
	}()

	message := map[string]interface{}{
		"username": "Splitoor",
		"embeds": []map[string]interface{}{
			{
				"title":       e.GetTitle(),
				"description": e.GetDescription(),
				"color":       16711680,
			},
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		errorType = "marshal_error"

		return fmt.Errorf("failed to marshal discord message: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.Webhook, bytes.NewBuffer(jsonData))
	if err != nil {
		errorType = "request_error"

		return fmt.Errorf("failed to create discord webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		errorType = "send_error"

		return fmt.Errorf("failed to send discord webhook: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		statusCode = strconv.Itoa(resp.StatusCode)
		errorType = "status_error"

		log.WithField("status_code", resp.StatusCode).Error("Discord webhook returned non-2xx status code")

		return fmt.Errorf("discord webhook returned non-2xx status code: %d", resp.StatusCode)
	}

	log.Info("Successfully published message to Discord")

	return nil
}
