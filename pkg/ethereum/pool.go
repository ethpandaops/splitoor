package ethereum

import (
	"context"
	"errors"
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
	metrics        *Metrics

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
		metrics:               GetMetricsInstance("splitoor_ethereum_pool", name),
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

	// Check for nil map before iterating
	if p.healthyExecutionNodes == nil {
		return healthyNodes
	}

	for node, healthy := range p.healthyExecutionNodes {
		// Skip nil nodes
		if node == nil {
			continue
		}

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

	// Check for nil map before iterating
	if p.healthyExecutionNodes == nil {
		return nil
	}

	for node, healthy := range p.healthyExecutionNodes {
		// Skip nil nodes
		if node == nil {
			continue
		}

		if healthy {
			healthyNodes = append(healthyNodes, node)
		}
	}

	if len(healthyNodes) == 0 {
		return nil
	}

	// Make sure we have a valid slice before accessing elements
	nodeCount := len(healthyNodes)
	if nodeCount == 0 {
		return nil
	}

	//nolint:gosec // doesn't matter
	randomIndex := rand.IntN(nodeCount)

	// Double check the index is valid
	if randomIndex < 0 || randomIndex >= nodeCount {
		return nil
	}

	return healthyNodes[randomIndex]
}

func (p *Pool) GetHealthyBeaconNodes() []*beacon.Node {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var healthyNodes []*beacon.Node

	// Check for nil map before iterating
	if p.healthyBeaconNodes == nil {
		return healthyNodes
	}

	for node, healthy := range p.healthyBeaconNodes {
		// Skip nil nodes
		if node == nil {
			continue
		}

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

	// Check for nil map before iterating
	if p.healthyBeaconNodes == nil {
		return nil
	}

	for node, healthy := range p.healthyBeaconNodes {
		// Skip nil nodes
		if node == nil {
			continue
		}

		if healthy {
			healthyNodes = append(healthyNodes, node)
		}
	}

	if len(healthyNodes) == 0 {
		return nil
	}

	// Make sure we have a valid slice before accessing elements
	nodeCount := len(healthyNodes)
	if nodeCount == 0 {
		return nil
	}

	//nolint:gosec // doesn't matter
	randomIndex := rand.IntN(nodeCount)

	// Double check the index is valid
	if randomIndex < 0 || randomIndex >= nodeCount {
		return nil
	}

	return healthyNodes[randomIndex]
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
	// Create a context with cancellation for proper resource management
	g, gCtx := errgroup.WithContext(ctx)

	// Use a separate waitgroup for tracking goroutines we spawn ourselves
	var wg sync.WaitGroup

	// Create a channel for error reporting from background goroutines
	errCh := make(chan error, 3) // Buffer size matches number of background goroutines

	// Initialize metrics first
	p.UpdateNodeMetrics()

	// Safely capture node references to avoid loop variable capture issues
	for _, beaconNode := range p.beaconNodes {
		// Create local copy of the loop variable to avoid data race
		node := beaconNode

		g.Go(func() error {
			// Check for nil node to avoid panic
			if node == nil {
				return fmt.Errorf("nil beacon node in pool")
			}

			node.OnReady(gCtx, func(ctx context.Context) error {
				p.mu.Lock()
				defer p.mu.Unlock()

				if p.healthyBeaconNodes == nil {
					p.healthyBeaconNodes = make(map[*beacon.Node]bool)
				}

				p.healthyBeaconNodes[node] = true

				return nil
			})

			return node.Start(gCtx)
		})
	}

	for _, execNode := range p.executionNodes {
		// Create local copy of the loop variable to avoid data race
		node := execNode

		g.Go(func() error {
			// Check for nil node to avoid panic
			if node == nil {
				return fmt.Errorf("nil execution node in pool")
			}

			node.OnReady(gCtx, func(ctx context.Context) error {
				p.mu.Lock()
				defer p.mu.Unlock()

				if p.healthyExecutionNodes == nil {
					p.healthyExecutionNodes = make(map[*execution.Node]bool)
				}

				p.healthyExecutionNodes[node] = true

				return nil
			})

			return node.Start(gCtx)
		})
	}

	// Start status reporting goroutine with proper context handling
	wg.Add(1)

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(1 * time.Minute)

		defer ticker.Stop()

		for {
			select {
			case <-gCtx.Done():
				return
			case <-ticker.C:
				func() {
					p.mu.RLock()
					defer p.mu.RUnlock()

					// Safely get counts, handling nil maps
					healthyBeacon := 0
					healthyExec := 0

					if p.healthyBeaconNodes != nil {
						healthyBeacon = len(p.healthyBeaconNodes)
					}

					if p.healthyExecutionNodes != nil {
						healthyExec = len(p.healthyExecutionNodes)
					}

					totalBeacon := len(p.beaconNodes)
					totalExec := len(p.executionNodes)

					p.log.WithFields(logrus.Fields{
						"healthy_beacon_nodes":    fmt.Sprintf("%d/%d", healthyBeacon, totalBeacon),
						"healthy_execution_nodes": fmt.Sprintf("%d/%d", healthyExec, totalExec),
					}).Info("Pool status")
				}()
			}
		}
	}()

	// Start metrics update goroutine
	wg.Add(1)

	go func() {
		defer wg.Done()

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

	// Handle errors from errgroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(errCh) // Close error channel when done

		if err := g.Wait(); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				// Context cancellation is expected, not an error
				return
			}

			// Report the error
			p.log.WithError(err).Error("error in node pool")

			// Send error to channel if context still valid
			select {
			case <-ctx.Done():
				return
			case errCh <- err:
			}
		}
	}()
}

func (p *Pool) UpdateNodeMetrics() {
	p.mu.Lock()

	// Safely get counts, handling nil maps
	healthyBeacon := 0
	healthyExec := 0

	if p.healthyBeaconNodes != nil {
		healthyBeacon = len(p.healthyBeaconNodes)
	}

	if p.healthyExecutionNodes != nil {
		healthyExec = len(p.healthyExecutionNodes)
	}

	totalBeacon := len(p.beaconNodes)
	totalExec := len(p.executionNodes)

	unhealthyBeacon := totalBeacon - healthyBeacon
	unhealthyExec := totalExec - healthyExec

	p.mu.Unlock()

	p.metrics.SetNodesTotal(float64(healthyBeacon), []string{"beacon", "healthy"})
	p.metrics.SetNodesTotal(float64(healthyExec), []string{"execution", "healthy"})
	p.metrics.SetNodesTotal(float64(unhealthyBeacon), []string{"beacon", "unhealthy"})
	p.metrics.SetNodesTotal(float64(unhealthyExec), []string{"execution", "unhealthy"})
}
