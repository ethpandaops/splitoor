package ses

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/ethpandaops/splitoor/pkg/monitor/event"
	"github.com/sirupsen/logrus"
)

type SES struct {
	log     logrus.FieldLogger
	monitor string
	name    string
	config  *Config
	client  *ses.SES
	metrics *Metrics
}

func NewSES(ctx context.Context, log logrus.FieldLogger, monitor, name string, config *Config) (*SES, error) {
	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &SES{
		log:     log.WithField("source", "ses"),
		monitor: monitor,
		name:    name,
		config:  config,
		client:  ses.New(sess),
		metrics: GetMetricsInstance("splitoor_notifier_ses", monitor),
	}, nil
}

func (s *SES) Start(ctx context.Context) error {
	return nil
}

func (s *SES) Stop(ctx context.Context) error {
	return nil
}

func (s *SES) GetType() string {
	return "ses"
}

func (s *SES) GetName() string {
	return s.name
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

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: aws.StringSlice(s.config.To),
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(e.GetDescription()),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(e.GetTitle()),
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
