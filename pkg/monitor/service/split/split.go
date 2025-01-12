package split

import (
	"context"

	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	"github.com/sirupsen/logrus"
)

type Service struct {
	log          logrus.FieldLogger
	config       *Config
	ethereumPool *ethereum.Pool
	publisher    *notifier.Publisher
}

func NewService(ctx context.Context, log logrus.FieldLogger, name string, config *Config, ethereumPool *ethereum.Pool, publisher *notifier.Publisher) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Service{
		log:          log.WithField("service", ServiceType),
		config:       config,
		ethereumPool: ethereumPool,
		publisher:    publisher,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	s.log.Info("Starting split service")
	// TODO: Implement monitoring logic for splits
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	s.log.Info("Stopping split service")

	return nil
}
