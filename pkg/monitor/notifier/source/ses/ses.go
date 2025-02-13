package ses

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/sirupsen/logrus"
)

const SourceType = "ses"

type SES struct {
	log     logrus.FieldLogger
	monitor string
	name    string
	config  *Config
	client  *ses.SES
	metrics *Metrics

	includeMonitorName bool
	includeGroupName   bool
	docs               *string
}

func NewSES(ctx context.Context, log logrus.FieldLogger, monitor, name string, docs *string, includeMonitorName, includeGroupName bool, config *Config) (*SES, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &SES{
		log:                log.WithField("source", "ses"),
		monitor:            monitor,
		name:               name,
		config:             config,
		client:             ses.New(sess),
		metrics:            GetMetricsInstance("splitoor_notifier_ses", monitor),
		includeMonitorName: includeMonitorName,
		includeGroupName:   includeGroupName,
		docs:               docs,
	}, nil
}

func (s *SES) Start(ctx context.Context) error {
	return nil
}

func (s *SES) Stop(ctx context.Context) error {
	return nil
}

func (s *SES) GetType() string {
	return SourceType
}

func (s *SES) GetName() string {
	return s.name
}

func (s *SES) GetConfig() *Config {
	return s.config
}

func (s *SES) Publish(ctx context.Context, e event.Event) error {
	var errorType string
	defer func() {
		if errorType != "" {
			s.metrics.IncErrors(s.monitor, s.name, s.GetType(), errorType)
		} else {
			s.metrics.IncMessagesPublished(s.monitor, s.name, s.GetType())
		}
	}()

	description := e.GetDescriptionText(s.includeMonitorName, s.includeGroupName)

	if s.docs != nil {
		docURL := strings.ReplaceAll(*s.docs, ":group", e.GetGroup())
		description = fmt.Sprintf("%s\n\nGo to docs: %s", description, docURL)
	}

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: aws.StringSlice(s.config.To),
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(description),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(fmt.Sprintf("🚨 %s", e.GetTitle(s.includeMonitorName, s.includeGroupName))),
			},
		},
		Source: aws.String(s.config.From),
	}

	_, err := s.client.SendEmailWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				errorType = "message_rejected"

				return fmt.Errorf("message rejected: %w", aerr)
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				errorType = "domain_not_verified"

				return fmt.Errorf("mail from domain not verified: %w", aerr)
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				errorType = "config_set_not_found"

				return fmt.Errorf("configuration set does not exist: %w", aerr)
			default:
				errorType = "unknown_aws_error"

				return fmt.Errorf("aws error: %w", aerr)
			}
		}

		errorType = "send_error"

		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
