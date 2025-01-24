package group

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/0xsequence/ethkit/ethcoder"
	"github.com/ethpandaops/splitoor/pkg/0xsplits/contract"
	spl "github.com/ethpandaops/splitoor/pkg/0xsplits/split"
	"github.com/ethpandaops/splitoor/pkg/ethereum"
	event "github.com/ethpandaops/splitoor/pkg/monitor/event/split"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	"github.com/ethpandaops/splitoor/pkg/monitor/safe"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/account"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/alert"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/controller"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	MinBalance                 uint64 = 31.95 * 1e9
	ExpectedStatuses                  = []string{"active_online", "active_offline"}
	WithdrawalCredentialsCodes        = []int64{1}
)

type Group struct {
	log     logrus.FieldLogger
	name    string
	monitor string

	publisher    *notifier.Publisher
	ethereumPool *ethereum.Pool
	safeClient   safe.Client

	address     string
	client      *spl.Client
	contractABI *ethcoder.ABI
	contract    string
	hash        string
	accounts    []*account.Account
	controller  controller.Controller

	metrics *Metrics

	hashAlert       *alert.Hash
	controllerAlert *alert.Controller
}

func NewGroup(ctx context.Context, log logrus.FieldLogger, monitor string, conf *Config, ethereumPool *ethereum.Pool, publisher *notifier.Publisher, safeClient safe.Client) (*Group, error) {
	log = log.WithField("group", conf.Name)

	var c string

	if conf.Contract != nil {
		c = *conf.Contract
	}

	accounts := make([]*account.Account, len(conf.Accounts))

	for i, acc := range conf.Accounts {
		accounts[i] = account.NewAccount(log, monitor, conf.Name, acc.Address, acc.Allocation, acc.Monitor, ethereumPool)
	}

	ctr, err := controller.NewController(ctx, log, monitor, conf.Name, conf.Controller.ControllerType, conf.Controller.Config, conf.Address, c, ethereumPool, safeClient, publisher)
	if err != nil {
		return nil, err
	}

	return &Group{
		log:             log.WithField("split", conf.Name),
		name:            conf.Name,
		monitor:         monitor,
		publisher:       publisher,
		ethereumPool:    ethereumPool,
		safeClient:      safeClient,
		address:         conf.Address,
		contract:        c,
		accounts:        accounts,
		controller:      ctr,
		metrics:         GetMetricsInstance("splitoor_split", monitor),
		hashAlert:       nil,
		controllerAlert: alert.NewController(log, ctr.Address()),
	}, nil
}

func (g *Group) Start(ctx context.Context) error {
	if g.controller != nil {
		if err := g.controller.Start(ctx); err != nil {
			return err
		}
	}

	if err := g.setupSplit(ctx); err != nil {
		return err
	}

	for _, account := range g.accounts {
		if err := account.Start(ctx); err != nil {
			return err
		}
	}

	g.tick(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 12):
				g.tick(ctx)
			}
		}
	}()

	return nil
}

func (g *Group) Stop(ctx context.Context) error {
	if g.controller != nil {
		if err := g.controller.Stop(ctx); err != nil {
			return err
		}
	}

	for _, account := range g.accounts {
		if err := account.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (g *Group) setupSplit(ctx context.Context) error {
	log := g.log.WithField("split", g.name)

	if g.contract == "" {
		log.Debug("no contract address provided for split, requesting default contract address")

		dpNode, err := g.ethereumPool.WaitForHealthyExecutionNode(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get healthy execution node to figure out default contract address")
		}

		chainID, err := dpNode.ChainID(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get chain id from execution node")
		}

		address := contract.GetDefaultContractAddress(chainID.String())
		if address == nil {
			return errors.New("failed to get default contract address for chain id " + chainID.String())
		}

		g.contract = *address
	}

	sCfg := &spl.Config{
		ContractAddress: g.contract,
		SplitAddress:    &g.address,
	}

	var err error

	g.client, err = spl.NewClient(log, sCfg)
	if err != nil {
		return errors.Wrap(err, "failed to create split client")
	}

	g.contractABI, err = contract.GetSplitMainAbi()
	if err != nil {
		return err
	}

	for _, account := range g.accounts {
		account.SetClient(g.client)
		account.SetContract(g.contractABI)
	}

	accounts := []string{}
	allocations := []uint32{}

	for _, account := range g.accounts {
		accounts = append(accounts, account.Address())
		allocations = append(allocations, account.Allocation())
	}

	hashParams := &spl.HashParams{
		Accounts:              accounts,
		PercentageAllocations: allocations,
	}

	hash, err := spl.CalculateHash(hashParams)
	if err != nil {
		return errors.Wrap(err, "failed to calculate hash")
	}

	g.hash = hex.EncodeToString(hash)

	g.hashAlert = alert.NewHash(g.log, g.hash)

	return nil
}

func (g *Group) tick(ctx context.Context) {
	go g.checkController(ctx)
	go g.checkHash(ctx)
	go g.gatherMetrics(ctx)
}

func (g *Group) checkController(ctx context.Context) {
	for _, node := range g.ethereumPool.GetHealthyExecutionNodes() {
		if ctx.Err() != nil {
			return
		}

		actualController, err := g.client.GetController(ctx, node, g.contractABI)
		if err != nil {
			g.log.WithError(err).Error("Error fetching controller")
		}

		val := float64(0)
		if *actualController == g.controller.Address() {
			val = 1
		}

		g.metrics.UpdateController(val, []string{g.name, node.Name(), g.address, g.controller.Address(), *actualController, g.controller.Type()})

		shouldAlert := g.controllerAlert.Update(*actualController)
		if shouldAlert {
			g.log.WithFields(logrus.Fields{
				"split_address":       g.address,
				"expected_controller": g.controller.Address(),
				"actual_controller":   *actualController,
			}).Warn("Alerting controller mismatch")

			if err := g.publisher.Publish(event.NewController(time.Now(), g.monitor, g.name, g.address, g.controller.Address(), *actualController)); err != nil {
				g.log.WithError(err).WithFields(logrus.Fields{
					"split_address":       g.address,
					"expected_controller": g.controller.Address(),
					"actual_controller":   *actualController,
				}).Error("Error publishing controller mismatch alert")
			}
		}
	}
}

func (g *Group) checkHash(ctx context.Context) {
	for _, node := range g.ethereumPool.GetHealthyExecutionNodes() {
		actualHash, err := g.client.GetHash(ctx, node, g.contractABI)
		if err != nil {
			g.log.WithError(err).Error("Error fetching hash")
		}

		if actualHash == nil {
			g.log.WithField("node", node.Name()).Error("Hash is nil")

			continue
		}

		actualHashString := hex.EncodeToString(actualHash[:])

		val := float64(0)
		if actualHashString == g.hash {
			val = 1
		}

		g.metrics.UpdateHash(val, []string{g.name, node.Name(), g.address, g.hash, actualHashString})

		shouldAlert := g.hashAlert.Update(actualHashString)
		if shouldAlert {
			g.log.WithFields(logrus.Fields{
				"split_address": g.address,
				"expected_hash": g.hash,
				"actual_hash":   actualHashString,
			}).Warn("Alerting hash mismatch")

			if err := g.publisher.Publish(event.NewHash(time.Now(), g.monitor, g.name, g.address, g.hash, actualHashString)); err != nil {
				g.log.WithError(err).WithFields(logrus.Fields{
					"split_address": g.address,
					"expected_hash": g.hash,
					"actual_hash":   actualHashString,
				}).Error("Error publishing hash mismatch alert")
			}
		}
	}
}

func (g *Group) gatherMetrics(ctx context.Context) {
	for _, node := range g.ethereumPool.GetHealthyExecutionNodes() {
		balance, err := node.BalanceAt(ctx, g.address)
		if err != nil {
			g.log.WithError(err).WithField("node", node.Name()).Error("Error fetching balance")
		}

		if balance == nil {
			g.log.WithField("node", node.Name()).Error("Balance is nil")

			continue
		}

		g.metrics.UpdateBalance(float64(balance.Uint64()), []string{g.name, node.Name(), g.address})
	}
}
