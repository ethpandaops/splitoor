package ethereum

import (
	"errors"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	nodesTotal *prometheus.GaugeVec
}

var (
	metricsInstances = make(map[string]*Metrics)
	metricsMutex     sync.Mutex
)

// GetMetricsInstance returns a metrics instance for the given namespace and monitor
// For production use, this will return a singleton for each namespace+monitor combination
// For testing, include a unique suffix in the namespace to avoid collisions.
func GetMetricsInstance(namespace, monitor string) *Metrics {
	key := fmt.Sprintf("%s-%s", namespace, monitor)

	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	if instance, exists := metricsInstances[key]; exists {
		return instance
	}

	// Create new metrics instance with the provided namespace
	constLabels := prometheus.Labels{"monitor": monitor}
	labels := []string{"type", "status"}

	// Create a metrics collector but register it only if it doesn't exist yet
	nodesTotal := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "nodes_total",
		Help:        "Total number of nodes in the pool",
		ConstLabels: constLabels,
	}, labels)

	// Try to register but don't panic if it fails (already registered)
	if err := prometheus.Register(nodesTotal); err != nil {
		// If it's already registered, try to find the existing one
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			// If we can recover the existing metric, use it
			if existing, isGauge := are.ExistingCollector.(*prometheus.GaugeVec); isGauge {
				nodesTotal = existing
			}
		}
	}

	instance := &Metrics{
		nodesTotal: nodesTotal,
	}

	metricsInstances[key] = instance

	return instance
}

func (m *Metrics) SetNodesTotal(count float64, labels []string) {
	if m == nil || m.nodesTotal == nil {
		return
	}

	m.nodesTotal.WithLabelValues(labels...).Set(count)
}
