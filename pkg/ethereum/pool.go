package ethereum

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/ethpandaops/splitoor/pkg/ethereum/beacon"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Pool struct {
	log            logrus.FieldLogger
	name           string
	beaconNodes    []*beacon.Node
	executionNodes []*execution.Node
	metrics        Metrics

	mu sync.RWMutex

	healthyBeaconNodes    map[*beacon.Node]bool
	healthyExecutionNodes map[*execution.Node]bool
}

func NewPool(ctx context.Context, log logrus.FieldLogger, name string, config *Config) *Pool {
	p := &Pool{
		log:                   log.WithField("module", "ethereum"),
		name:                  name,
		beaconNodes:           make([]*beacon.Node, 0),
		executionNodes:        make([]*execution.Node, 0),
		healthyBeaconNodes:    make(map[*beacon.Node]bool),
		healthyExecutionNodes: make(map[*execution.Node]bool),
		metrics:               NewMetrics("splitoor_ethereum_pool", name),
	}

	for i, beaconCfg := range config.Beacon {
		node := beacon.NewNode(ctx, log, fmt.Sprintf("beacon-%d", i), beaconCfg)
		p.beaconNodes = append(p.beaconNodes, node)
	}

	for i, execCfg := range config.Execution {
		node := execution.NewNode(log, fmt.Sprintf("execution-%d", i), execCfg)
		p.executionNodes = append(p.executionNodes, node)
	}

	return p
}

func (p *Pool) HasExecutionNodes() bool {
	return len(p.executionNodes) > 0
}

func (p *Pool) HasBeaconNodes() bool {
	return len(p.beaconNodes) > 0
}

func (p *Pool) HasHealthyBeaconNodes() bool {
	return len(p.healthyBeaconNodes) > 0
}

func (p *Pool) HasHealthyExecutionNodes() bool {
	return len(p.healthyExecutionNodes) > 0
}

func (p *Pool) GetHealthyExecutionNodes() []*execution.Node {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var healthyNodes []*execution.Node

	for node, healthy := range p.healthyExecutionNodes {
		if healthy {
			healthyNodes = append(healthyNodes, node)
		}
	}

	return healthyNodes
}

func (p *Pool) GetHealthyExecutionNode() *execution.Node {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var healthyNodes []*execution.Node

	for node, healthy := range p.healthyExecutionNodes {
		if healthy {
			healthyNodes = append(healthyNodes, node)
		}
	}

	if len(healthyNodes) == 0 {
		return nil
	}

	//nolint:gosec // doesn't matter
	return healthyNodes[rand.IntN(len(healthyNodes))]
}

func (p *Pool) GetHealthyBeaconNodes() []*beacon.Node {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var healthyNodes []*beacon.Node

	for node, healthy := range p.healthyBeaconNodes {
		if healthy {
			healthyNodes = append(healthyNodes, node)
		}
	}

	return healthyNodes
}

func (p *Pool) GetHealthyBeaconNode() *beacon.Node {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var healthyNodes []*beacon.Node

	for node, healthy := range p.healthyBeaconNodes {
		if healthy {
			healthyNodes = append(healthyNodes, node)
		}
	}

	if len(healthyNodes) == 0 {
		return nil
	}
	//nolint:gosec // doesn't matter
	return healthyNodes[rand.IntN(len(healthyNodes))]
}

func (p *Pool) WaitForHealthyBeaconNode(ctx context.Context) (*beacon.Node, error) {
	for {
		if node := p.GetHealthyBeaconNode(); node != nil {
			return node, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func (p *Pool) WaitForHealthyExecutionNode(ctx context.Context) (*execution.Node, error) {
	for {
		if node := p.GetHealthyExecutionNode(); node != nil {
			return node, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func (p *Pool) Start(ctx context.Context) {
	g, gCtx := errgroup.WithContext(ctx)

	p.UpdateNodeMetrics()

	for _, node := range p.beaconNodes {
		g.Go(func() error {
			node.OnReady(gCtx, func(ctx context.Context) error {
				p.mu.Lock()
				p.healthyBeaconNodes[node] = true
				p.mu.Unlock()

				return nil
			})

			return node.Start(gCtx)
		})
	}

	for _, node := range p.executionNodes {
		g.Go(func() error {
			node.OnReady(gCtx, func(ctx context.Context) error {
				p.mu.Lock()
				p.healthyExecutionNodes[node] = true
				p.mu.Unlock()

				return nil
			})

			return node.Start(gCtx)
		})
	}

	// Start status reporting goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-gCtx.Done():
				return
			case <-ticker.C:
				p.mu.RLock()
				healthyBeacon := len(p.healthyBeaconNodes)
				healthyExec := len(p.healthyExecutionNodes)
				totalBeacon := len(p.beaconNodes)
				totalExec := len(p.executionNodes)
				p.mu.RUnlock()

				p.log.WithFields(logrus.Fields{
					"healthy_beacon_nodes":    fmt.Sprintf("%d/%d", healthyBeacon, totalBeacon),
					"healthy_execution_nodes": fmt.Sprintf("%d/%d", healthyExec, totalExec),
				}).Info("Pool status")
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-gCtx.Done():
				return
			case <-ticker.C:
				p.UpdateNodeMetrics()
			}
		}
	}()

	go func() {
		if err := g.Wait(); err != nil {
			if ctx.Err() != nil {
				return
			}

			p.log.WithError(err).Error("error in pool")
		}
	}()
}

func (p *Pool) UpdateNodeMetrics() {
	p.mu.Lock()
	healthyBeacon := len(p.healthyBeaconNodes)
	healthyExec := len(p.healthyExecutionNodes)
	unhealthyBeacon := len(p.beaconNodes) - healthyBeacon
	unhealthyExec := len(p.executionNodes) - healthyExec
	p.mu.Unlock()

	p.metrics.SetNodesTotal(float64(healthyBeacon), []string{"beacon", "healthy"})
	p.metrics.SetNodesTotal(float64(healthyExec), []string{"execution", "healthy"})
	p.metrics.SetNodesTotal(float64(unhealthyBeacon), []string{"beacon", "unhealthy"})
	p.metrics.SetNodesTotal(float64(unhealthyExec), []string{"execution", "unhealthy"})
}
