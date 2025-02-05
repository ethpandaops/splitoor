package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/sirupsen/logrus"
)

const SourceType = "smtp"

type SMTP struct {
	log      logrus.FieldLogger
	config   *Config
	monitor  string
	name     string
	smtpAuth smtp.Auth
	metrics  *Metrics

	includeMonitorName bool
	includeGroupName   bool
	docs               *string
}

func NewSMTP(ctx context.Context, log logrus.FieldLogger, monitor, name string, docs *string, includeMonitorName, includeGroupName bool, config *Config) (*SMTP, error) {
	var auth smtp.Auth
	if config.Username != "" {
		auth = smtp.PlainAuth("", config.Username, config.Password, config.Host)
	}

	return &SMTP{
		log:                log.WithField("source", "smtp"),
		config:             config,
		monitor:            monitor,
		name:               name,
		smtpAuth:           auth,
		metrics:            GetMetricsInstance("splitoor_notifier_smtp", monitor),
		includeMonitorName: includeMonitorName,
		includeGroupName:   includeGroupName,
		docs:               docs,
	}, nil
}

func (s *SMTP) Start(ctx context.Context) error {
	s.log.Info("Starting SMTP source")

	return nil
}

func (s *SMTP) Stop(ctx context.Context) error {
	s.log.Info("Stopping SMTP source")

	return nil
}

func (s *SMTP) GetType() string {
	return SourceType
}

func (s *SMTP) GetName() string {
	return s.name
}

func (s *SMTP) GetConfig() *Config {
	return s.config
}

func (s *SMTP) Publish(ctx context.Context, evt event.Event) error {
	description := evt.GetDescriptionText(s.includeMonitorName, s.includeGroupName)

	if s.docs != nil {
		docURL := strings.ReplaceAll(*s.docs, ":group", evt.GetGroup())
		description = fmt.Sprintf("%s\n\nGo to docs: %s", description, docURL)
	}

	if err := s.sendEmail(evt, fmt.Sprintf("🚨 %s", evt.GetTitle(s.includeMonitorName, s.includeGroupName)), description); err != nil {
		return err
	}

	s.metrics.IncMessagesPublished(evt.GetGroup(), s.name, s.GetType())

	return nil
}

func (s *SMTP) sendEmail(evt event.Event, subject, body string) error {
	msg := fmt.Sprintf("From: %s\r\n"+
		"Bcc: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", s.config.From, strings.Join(s.config.To, ","), subject, body)

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	var errorType string
	defer func() {
		if errorType != "" {
			s.metrics.IncErrors(evt.GetGroup(), s.name, s.GetType(), errorType)
		}
	}()

	s.log.WithField("tls", s.config.TLS).Info("Sending email")

	if s.config.TLS {
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
			MinVersion: tls.VersionTLS12,
			//nolint:gosec // InsecureSkipVerify is configurable by the user
			InsecureSkipVerify: s.config.InsecureSkipVerify,
		}

		client, err := smtp.Dial(addr)
		if err != nil {
			errorType = "dial_error"

			return fmt.Errorf("failed to dial SMTP server: %w", err)
		}
		defer client.Close()

		if tlsErr := client.StartTLS(tlsConfig); tlsErr != nil {
			errorType = "tls_error"

			return fmt.Errorf("failed to start TLS: %w", tlsErr)
		}

		if s.smtpAuth != nil {
			if authErr := client.Auth(s.smtpAuth); authErr != nil {
				errorType = "auth_error"

				return fmt.Errorf("failed to authenticate: %w", authErr)
			}
		}

		if mailErr := client.Mail(s.config.From); mailErr != nil {
			errorType = "sender_error"

			return fmt.Errorf("failed to set sender: %w", mailErr)
		}

		for _, to := range s.config.To {
			if rcptErr := client.Rcpt(to); rcptErr != nil {
				errorType = "recipient_error"

				return fmt.Errorf("failed to add recipient %s: %w", to, rcptErr)
			}
		}

		w, err := client.Data()
		if err != nil {
			errorType = "data_error"

			return fmt.Errorf("failed to create message writer: %w", err)
		}
		defer w.Close()

		_, err = w.Write([]byte(msg))
		if err != nil {
			errorType = "write_error"

			return fmt.Errorf("failed to write message: %w", err)
		}

		return nil
	}

	if err := smtp.SendMail(addr, s.smtpAuth, s.config.From, s.config.To, []byte(msg)); err != nil {
		errorType = "send_error"

		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
