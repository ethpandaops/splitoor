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

	address         string
	recoveryAddress string

	client       *spl.Client
	contractABI  *ethcoder.ABI
	contract     string
	initialHash  string
	recoveryHash string
	stableHash   string
	accounts     []*account.Account
	controller   controller.Controller

	metrics *Metrics

	hashUnknownAlert  *alert.HashUnknown
	hashInitialAlert  *alert.HashInitial
	hashRecoveryAlert *alert.HashRecovery
	controllerAlert   *alert.Controller
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

	ctr, err := controller.NewController(ctx, log, monitor, conf.Name, conf.Controller.ControllerType, conf.Controller.Config, conf.Address, conf.RecoveryAddress, c, ethereumPool, safeClient, publisher)
	if err != nil {
		return nil, err
	}

	return &Group{
		log:               log.WithField("split", conf.Name),
		name:              conf.Name,
		monitor:           monitor,
		publisher:         publisher,
		ethereumPool:      ethereumPool,
		safeClient:        safeClient,
		address:           conf.Address,
		recoveryAddress:   conf.RecoveryAddress,
		contract:          c,
		accounts:          accounts,
		controller:        ctr,
		metrics:           GetMetricsInstance("splitoor_split", monitor),
		hashUnknownAlert:  nil,
		hashInitialAlert:  nil,
		hashRecoveryAlert: nil,
		controllerAlert:   alert.NewController(log, ctr.Address()),
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

	// stable hash accounts and allocations
	accounts := []string{}
	allocations := []uint32{}

	for _, account := range g.accounts {
		accounts = append(accounts, account.Address())
		allocations = append(allocations, account.Allocation())
	}

	// Calculate the stable hash of the split
	g.stableHash, err = calculateHash(accounts, allocations)
	if err != nil {
		return errors.Wrap(err, "failed to calculate stable hash")
	}

	// initial hash accounts and allocations
	g.initialHash, err = calculateHash([]string{
		g.recoveryAddress,
		g.controller.Address(),
	}, []uint32{
		999999,
		1,
	})
	if err != nil {
		return errors.Wrap(err, "failed to calculate initial hash")
	}

	// recovery hash accounts and allocations
	g.recoveryHash, err = calculateHash([]string{
		g.address,
		g.recoveryAddress,
	}, []uint32{
		1,
		999999,
	})
	if err != nil {
		return errors.Wrap(err, "failed to calculate recovery hash")
	}

	g.hashUnknownAlert = alert.NewHashUnknown(g.log, []string{g.initialHash, g.stableHash, g.recoveryHash})
	g.hashInitialAlert = alert.NewHashInitial(g.log, g.initialHash)
	g.hashRecoveryAlert = alert.NewHashRecovery(g.log, g.recoveryHash)

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

			continue
		}

		if actualController == nil {
			g.log.WithField("node", node.Name()).Error("Controller is nil")
			g.metrics.UpdateController(0, []string{g.name, node.Name(), g.address, g.controller.Address(), "nil", g.controller.Type()})

			continue
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

		stableHashVal := float64(0)
		if actualHashString == g.stableHash {
			stableHashVal = 1
		}

		initialHashVal := float64(0)
		if actualHashString == g.initialHash {
			initialHashVal = 1
		}

		recoveryHashVal := float64(0)
		if actualHashString == g.recoveryHash {
			recoveryHashVal = 1
		}

		g.metrics.UpdateHashStable(stableHashVal, []string{g.name, node.Name(), g.address, g.stableHash, actualHashString})
		g.metrics.UpdateHashInitial(initialHashVal, []string{g.name, node.Name(), g.address, g.initialHash, actualHashString})
		g.metrics.UpdateHashRecovery(recoveryHashVal, []string{g.name, node.Name(), g.address, g.recoveryHash, actualHashString})

		shouldAlertUnknown := g.hashUnknownAlert.Update(actualHashString)
		if shouldAlertUnknown {
			g.log.WithFields(logrus.Fields{
				"split_address": g.address,
				"expected_hash": g.stableHash,
				"actual_hash":   actualHashString,
			}).Warn("Alerting stable hash unknown")

			if err := g.publisher.Publish(event.NewHashUnknownState(time.Now(), g.monitor, g.name, g.address, g.stableHash, actualHashString)); err != nil {
				g.log.WithError(err).WithFields(logrus.Fields{
					"split_address": g.address,
					"expected_hash": g.stableHash,
					"actual_hash":   actualHashString,
				}).Error("Error publishing hash unknown alert")
			}
		}

		shouldAlertInitial := g.hashInitialAlert.Update(actualHashString)
		if shouldAlertInitial {
			g.log.WithFields(logrus.Fields{
				"split_address": g.address,
				"expected_hash": g.initialHash,
				"actual_hash":   actualHashString,
			}).Warn("Alerting in initial hash state")

			if err := g.publisher.Publish(event.NewHashInitialState(time.Now(), g.monitor, g.name, g.address, actualHashString)); err != nil {
				g.log.WithError(err).WithFields(logrus.Fields{
					"split_address": g.address,
					"expected_hash": g.initialHash,
					"actual_hash":   actualHashString,
				}).Error("Error publishing in initial hash state alert")
			}
		}

		shouldAlertRecovery := g.hashRecoveryAlert.Update(actualHashString)
		if shouldAlertRecovery {
			g.log.WithFields(logrus.Fields{
				"split_address": g.address,
				"expected_hash": g.recoveryHash,
				"actual_hash":   actualHashString,
			}).Warn("Alerting in recovery hash state")

			if err := g.publisher.Publish(event.NewHashRecoveryState(time.Now(), g.monitor, g.name, g.address, actualHashString)); err != nil {
				g.log.WithError(err).WithFields(logrus.Fields{
					"split_address": g.address,
					"expected_hash": g.recoveryHash,
					"actual_hash":   actualHashString,
				}).Error("Error publishing in recovery hash state alert")
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
