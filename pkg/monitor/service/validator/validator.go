package validator

import (
	"context"
	"fmt"

	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/beaconchain"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/validator/group"
	"github.com/sirupsen/logrus"
)

type Service struct {
	log          logrus.FieldLogger
	config       *Config
	ethereumPool *ethereum.Pool
	publisher    *notifier.Publisher
	beaconchain  beaconchain.Client
	groups       []*group.Group
}

func NewService(ctx context.Context, log logrus.FieldLogger, monitor string, config *Config, ethereumPool *ethereum.Pool, publisher *notifier.Publisher, beaconchainClient beaconchain.Client) (*Service, error) {
	if beaconchainClient == nil && !ethereumPool.HasBeaconNodes() {
		return nil, fmt.Errorf("no beaconchain client or ethereum beacon nodes configured")
	}

	groups := make([]*group.Group, 0, len(config.Groups))

	for _, g := range config.Groups {
		ng, err := group.NewGroup(ctx, log, monitor, &g, ethereumPool, beaconchainClient, publisher)
		if err != nil {
			return nil, fmt.Errorf("failed to create group client: %w", err)
		}

		groups = append(groups, ng)
	}

	return &Service{
		log:          log.WithField("service", ServiceType),
		config:       config,
		ethereumPool: ethereumPool,
		publisher:    publisher,
		beaconchain:  beaconchainClient,
		groups:       groups,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	s.log.Info("Starting validator service")

	for _, g := range s.groups {
		go g.Start(ctx)
	}

	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	s.log.Info("Stopping validator service")

	for _, g := range s.groups {
		if err := g.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}
