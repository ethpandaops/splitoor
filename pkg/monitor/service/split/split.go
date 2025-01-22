package split

import (
	"context"

	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	"github.com/ethpandaops/splitoor/pkg/monitor/safe"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Service struct {
	log          logrus.FieldLogger
	config       *Config
	ethereumPool *ethereum.Pool
	publisher    *notifier.Publisher
	safeClient   safe.Client

	groups []*group.Group
}

func NewService(ctx context.Context, log logrus.FieldLogger, monitor string, config *Config, ethereumPool *ethereum.Pool, publisher *notifier.Publisher, safeClient safe.Client) (*Service, error) {
	groups := make([]*group.Group, len(config.Groups))

	for i, g := range config.Groups {
		ng, err := group.NewGroup(ctx, log, monitor, &g, ethereumPool, publisher, safeClient)
		if err != nil {
			return nil, err
		}

		groups[i] = ng
	}

	return &Service{
		log:          log.WithField("service", ServiceType),
		config:       config,
		ethereumPool: ethereumPool,
		publisher:    publisher,
		safeClient:   safeClient,
		groups:       groups,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	s.log.Info("Starting split service")

	if s.safeClient != nil {
		dpNode, err := s.ethereumPool.WaitForHealthyExecutionNode(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get healthy execution node to figure out default contract address")
		}

		chainID, err := dpNode.ChainID(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get chain id from execution node")
		}

		s.safeClient.SetChainID(chainID.String())
	}

	for _, g := range s.groups {
		if err := g.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	s.log.Info("Stopping split service")

	for _, g := range s.groups {
		if err := g.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}
