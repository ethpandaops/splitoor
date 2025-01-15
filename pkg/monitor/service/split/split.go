package split

import (
	"context"
	"encoding/hex"

	"github.com/ethpandaops/splitoor/pkg/0xsplits/contract"
	spl "github.com/ethpandaops/splitoor/pkg/0xsplits/split"
	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Service struct {
	log          logrus.FieldLogger
	config       *Config
	ethereumPool *ethereum.Pool
	publisher    *notifier.Publisher

	splits []*SplitItem
}

type SplitItem struct {
	Split
	client *spl.Client
	Hash   string
}

func NewService(ctx context.Context, log logrus.FieldLogger, name string, config *Config, ethereumPool *ethereum.Pool, publisher *notifier.Publisher) (*Service, error) {
	return &Service{
		log:          log.WithField("service", ServiceType),
		config:       config,
		ethereumPool: ethereumPool,
		publisher:    publisher,
		splits:       make([]*SplitItem, 0),
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	s.log.Info("Starting split service")

	for _, split := range s.config.Splits {
		splitItem, err := s.setupSplit(ctx, &split)
		if err != nil {
			return err
		}

		s.splits = append(s.splits, splitItem)
	}

	return nil
}

func (s *Service) setupSplit(ctx context.Context, split *Split) (*SplitItem, error) {
	log := s.log.WithField("split", split.Name)

	contractAddress := split.Contract

	if contractAddress == "" {
		log.Debugf("no contract address provided for split %s, requesting default contract address", split.Name)

		dpNode, err := s.ethereumPool.WaitForHealthyExecutionNode(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get healthy execution node to figure out default contract address")
		}

		chainID, err := dpNode.ChainID(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get chain id from execution node")
		}

		address := contract.GetDefaultContractAddress(chainID.String())
		if address == nil {
			return nil, errors.New("failed to get default contract address for chain id " + chainID.String())
		}

		contractAddress = *address
	}

	sCfg := &spl.Config{
		ContractAddress: contractAddress,
		SplitAddress:    &split.Address,
	}

	splitClient, err := spl.NewClient(log, sCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create split client for split "+split.Name)
	}

	accounts := []string{}
	allocations := []uint32{}

	for _, account := range split.Accounts {
		accounts = append(accounts, account.Address)
		allocations = append(allocations, uint32(account.Allocation))
	}

	hashParams := &spl.HashParams{
		Accounts:              accounts,
		PercentageAllocations: allocations,
	}

	hash, err := splitClient.CalculateHash(hashParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate hash for split "+split.Name)
	}

	return &SplitItem{
		Hash:   hex.EncodeToString(hash),
		Split:  *split,
		client: splitClient,
	}, nil
}

func (s *Service) Stop(ctx context.Context) error {
	s.log.Info("Stopping split service")

	return nil
}
