package eoa

import (
	"context"
	"math/big"
	"time"

	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/sirupsen/logrus"
)

const ControllerType = "eoa"

type EOA struct {
	log          logrus.FieldLogger
	name         string
	monitor      string
	address      string
	ethereumPool *ethereum.Pool

	metrics *Metrics
}

func New(ctx context.Context, log logrus.FieldLogger, monitor, name string, config *Config, ethereumPool *ethereum.Pool) (*EOA, error) {
	return &EOA{
		log:          log,
		name:         name,
		monitor:      monitor,
		address:      config.Address,
		ethereumPool: ethereumPool,
		metrics:      GetMetricsInstance("splitoor_split_controller", monitor),
	}, nil
}

func (c *EOA) Start(ctx context.Context) error {
	c.tick(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 12):
				c.tick(ctx)
			}
		}
	}()

	return nil
}

func (c *EOA) Stop(ctx context.Context) error {
	return nil
}

func (c *EOA) Type() string {
	return ControllerType
}

func (c *EOA) Name() string {
	return c.name
}

func (c *EOA) Address() string {
	return c.address
}

func (c *EOA) tick(ctx context.Context) {
	go c.gatherMetrics(ctx)
}

func (c *EOA) gatherMetrics(ctx context.Context) {
	for _, node := range c.ethereumPool.GetHealthyExecutionNodes() {
		if ctx.Err() != nil {
			return
		}

		balance, err := node.BalanceAt(ctx, c.address)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			c.log.WithError(err).WithField("node", node.Name()).Error("Error fetching balance")
		}

		if balance == nil {
			c.log.WithField("node", node.Name()).Error("Balance is nil")

			continue
		}

		balanceFloat := new(big.Float).SetInt(balance)
		balanceFloat64, _ := balanceFloat.Float64()

		c.metrics.UpdateBalance(balanceFloat64, []string{c.name, c.Type(), node.Name(), c.address})
	}
}
