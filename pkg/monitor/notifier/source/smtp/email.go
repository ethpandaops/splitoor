package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/sirupsen/logrus"
)

type SMTP struct {
	log      logrus.FieldLogger
	config   *Config
	monitor  string
	name     string
	smtpAuth smtp.Auth
	metrics  *Metrics
}

func NewSMTP(ctx context.Context, log logrus.FieldLogger, monitor, name string, config *Config) (*SMTP, error) {
	var auth smtp.Auth
	if config.Username != "" {
		auth = smtp.PlainAuth("", config.Username, config.Password, config.Host)
	}

	return &SMTP{
		log:      log.WithField("source", "smtp"),
		config:   config,
		monitor:  monitor,
		name:     name,
		smtpAuth: auth,
		metrics:  GetMetricsInstance("splitoor_notifier_smtp", monitor),
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
	return "smtp"
}

func (s *SMTP) GetName() string {
	return s.name
}

func (s *SMTP) Publish(ctx context.Context, evt event.Event) error {
	subject := fmt.Sprintf("[%s] %s Alert", s.monitor, evt.GetType())
	body := fmt.Sprintf("Event Type: %s\nGroup: %s\n\n%s", evt.GetType(), evt.GetGroup(), evt.GetText())

	if err := s.sendEmail(evt, subject, body); err != nil {
		return err
	}

	s.metrics.IncMessagesPublished(evt.GetGroup(), s.name, s.GetType())
	return nil
}

func (s *SMTP) sendEmail(evt event.Event, subject, body string) error {
	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", s.config.From, s.config.To[0], subject, body)

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	var errorType string
	defer func() {
		if errorType != "" {
			s.metrics.IncErrors(evt.GetGroup(), s.name, s.GetType(), errorType)
		}
	}()

	if s.config.TLS {
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
		}

		client, err := smtp.Dial(addr)
		if err != nil {
			errorType = "dial_error"
			return fmt.Errorf("failed to dial SMTP server: %w", err)
		}
		defer client.Close()

		if err := client.StartTLS(tlsConfig); err != nil {
			errorType = "tls_error"
			return fmt.Errorf("failed to start TLS: %w", err)
		}

		if s.smtpAuth != nil {
			if err := client.Auth(s.smtpAuth); err != nil {
				errorType = "auth_error"
				return fmt.Errorf("failed to authenticate: %w", err)
			}
		}

		if err := client.Mail(s.config.From); err != nil {
			errorType = "sender_error"
			return fmt.Errorf("failed to set sender: %w", err)
		}

		for _, to := range s.config.To {
			if err := client.Rcpt(to); err != nil {
				errorType = "recipient_error"
				return fmt.Errorf("failed to add recipient %s: %w", to, err)
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
