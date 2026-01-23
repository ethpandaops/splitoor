package ethereum

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ethpandaops/splitoor/pkg/ethereum/beacon"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Test the pool's behavior with nil maps and nodes.
func TestPoolSafety(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	tests := []struct {
		name           string
		beaconNodes    []*beacon.Node
		executionNodes []*execution.Node
	}{
		{
			name:           "nil_maps",
			beaconNodes:    nil,
			executionNodes: nil,
		},
		{
			name:           "empty_nodes",
			beaconNodes:    []*beacon.Node{},
			executionNodes: []*execution.Node{},
		},
		{
			name:           "nil_node_in_list",
			beaconNodes:    []*beacon.Node{nil},
			executionNodes: []*execution.Node{nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a unique metrics name for each test run to avoid collisions
			metricName := fmt.Sprintf("test_pool_safety_%d", time.Now().UnixNano())

			// Create pool with custom nodes for testing edge cases
			p := &Pool{
				log:                   log,
				name:                  "test",
				beaconNodes:           tt.beaconNodes,
				executionNodes:        tt.executionNodes,
				healthyBeaconNodes:    nil, // Deliberately nil to test safety
				healthyExecutionNodes: nil, // Deliberately nil to test safety
				metrics:               GetMetricsInstance(metricName, "test"),
			}

			// This tests that the GetHealthyExecutionNodes doesn't panic with nil maps
			execNodes := p.GetHealthyExecutionNodes()
			// The actual implementation returns an empty (non-nil) slice
			// which is the correct behavior, but we're actually testing that it doesn't panic
			if execNodes != nil {
				assert.Len(t, execNodes, 0)
			}

			// This tests that the GetHealthyBeaconNodes doesn't panic with nil maps
			beaconNodes := p.GetHealthyBeaconNodes()
			// The actual implementation returns an empty (non-nil) slice
			// which is the correct behavior, but we're actually testing that it doesn't panic
			if beaconNodes != nil {
				assert.Len(t, beaconNodes, 0)
			}

			// Test GetHealthyExecutionNode with nil maps
			execNode := p.GetHealthyExecutionNode()
			assert.Nil(t, execNode)

			// Test GetHealthyBeaconNode with nil maps
			beaconNode := p.GetHealthyBeaconNode()
			assert.Nil(t, beaconNode)

			// Test metrics doesn't panic
			p.UpdateNodeMetrics()
		})
	}
}

// Test concurrent access to the pool.
func TestPoolConcurrentAccess(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	// Use a unique name to avoid metrics collision
	metricName := fmt.Sprintf("test_pool_concurrent_%d", time.Now().UnixNano())

	// Create a pool with some actual data
	p := &Pool{
		log:                   log,
		name:                  "test",
		beaconNodes:           make([]*beacon.Node, 0),
		executionNodes:        make([]*execution.Node, 0),
		healthyBeaconNodes:    make(map[*beacon.Node]bool),
		healthyExecutionNodes: make(map[*execution.Node]bool),
		metrics:               GetMetricsInstance(metricName, "test"),
	}

	// Add a dummy execution node
	execNode := execution.NewNode(log, "test-exec", &execution.Config{})
	p.executionNodes = append(p.executionNodes, execNode)

	// Test concurrent access to the pool maps
	var wg sync.WaitGroup

	// Create a context with timeout
	ctx := context.Background()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	// Run multiple goroutines that access the pool concurrently
	for i := 0; i < 10; i++ {
		wg.Add(5)

		// Test reader methods
		go func() {
			defer wg.Done()

			_ = p.GetHealthyExecutionNodes()
		}()

		go func() {
			defer wg.Done()

			_ = p.GetHealthyBeaconNodes()
		}()

		go func() {
			defer wg.Done()

			_ = p.GetHealthyExecutionNode()
		}()

		go func() {
			defer wg.Done()

			_ = p.GetHealthyBeaconNode()
		}()

		// Test writer method with read lock contention
		go func() {
			defer wg.Done()

			p.UpdateNodeMetrics()
		}()
	}

	// Wait with timeout to ensure the test doesn't hang
	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for either completion or timeout
	select {
	case <-done:
		// Success - all goroutines completed without deadlock
	case <-ctxWithTimeout.Done():
		t.Fatalf("Test timed out, possible deadlock detected")
	}
}

// Test that context cancellation is handled properly.
func TestPoolContextCancellation(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	// Use a unique name to avoid metrics collision
	metricName := fmt.Sprintf("test_pool_cancel_%d", time.Now().UnixNano())

	config := &Config{
		Beacon:    []*beacon.Config{{}},
		Execution: []*execution.Config{{}},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create a new pool with the unique metrics name
	p := NewPool(ctx, log, "test", config)

	// Override the metrics with our unique instance
	p.metrics = GetMetricsInstance(metricName, "test")

	// Start the pool
	p.Start(ctx)

	// Immediately cancel to test cleanup
	cancel()

	// Allow some time for goroutines to exit
	time.Sleep(100 * time.Millisecond)

	// We can't easily assert that goroutines have exited, but this test
	// verifies the code runs without panics on context cancellation
}
