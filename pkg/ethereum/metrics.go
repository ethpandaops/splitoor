package ethereum

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	nodesTotal *prometheus.GaugeVec
}

func NewMetrics(namespace, monitorName string) Metrics {
	constLabels := prometheus.Labels{"monitor": monitorName}
	labels := []string{"type", "status"}

	m := Metrics{
		nodesTotal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "nodes_total",
			Help:        "Total number of nodes in the pool",
			ConstLabels: constLabels,
		}, labels),
	}

	prometheus.MustRegister(m.nodesTotal)

	return m
}

func (m Metrics) SetNodesTotal(count float64, labels []string) {
	m.nodesTotal.WithLabelValues(labels...).Set(count)
}
