package account

import (
	"context"
	"time"

	"github.com/0xsequence/ethkit/ethcoder"
	spl "github.com/ethpandaops/splitoor/pkg/0xsplits/split"
	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/sirupsen/logrus"
)

type Account struct {
	log                 logrus.FieldLogger
	name                string
	monitor             string
	address             string
	allocation          uint32
	shouldGatherMetrics bool
	ethereumPool        *ethereum.Pool

	client *spl.Client

	contract *ethcoder.ABI

	metrics *Metrics
}

func NewAccount(log logrus.FieldLogger, monitor, name, address string, allocation uint32, shouldGatherMetrics bool, ethereumPool *ethereum.Pool) *Account {
	return &Account{
		log:                 log.WithField("account", address),
		monitor:             monitor,
		name:                name,
		address:             address,
		allocation:          allocation,
		shouldGatherMetrics: shouldGatherMetrics,
		ethereumPool:        ethereumPool,
		metrics:             GetMetricsInstance("splitoor_split_account", monitor),
	}
}

func (a *Account) SetClient(client *spl.Client) {
	a.client = client
}

func (a *Account) SetContract(contract *ethcoder.ABI) {
	a.contract = contract
}

func (a *Account) Start(ctx context.Context) error {
	if !a.shouldGatherMetrics {
		return nil
	}

	a.tick(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 12):
				a.tick(ctx)
			}
		}
	}()

	return nil
}

func (a *Account) Stop(ctx context.Context) error {
	return nil
}

func (a *Account) Address() string {
	return a.address
}

func (a *Account) Allocation() uint32 {
	return a.allocation
}

func (a *Account) tick(ctx context.Context) {
	for _, node := range a.ethereumPool.GetHealthyExecutionNodes() {
		balance, err := node.BalanceAt(ctx, a.address)
		if err != nil {
			a.log.WithError(err).WithField("node", node.Name()).Error("Error fetching balance")
		}

		a.metrics.UpdateBalance(float64(balance.Uint64()), []string{a.name, node.Name(), a.address})

		if a.client != nil && a.contract != nil {
			balance, err := a.client.GetETHBalance(ctx, node, a.contract, a.address)
			if err != nil {
				a.log.WithError(err).WithField("node", node.Name()).Error("Error fetching split account balance")
			}

			a.metrics.UpdateSplitBalance(float64(balance.Uint64()), []string{a.name, node.Name(), a.address})
		}
	}
}
