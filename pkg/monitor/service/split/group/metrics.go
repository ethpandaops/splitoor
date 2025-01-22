package group

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	balance    *prometheus.GaugeVec
	hash       *prometheus.GaugeVec
	controller *prometheus.GaugeVec
}

var (
	metricsInstance *Metrics
	once            sync.Once
)

func GetMetricsInstance(namespace, monitor string) *Metrics {
	once.Do(func() {
		constLabels := prometheus.Labels{"monitor": monitor}

		metricsInstance = &Metrics{
			balance: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "balance",
					Help:        "The balance of the split.",
					ConstLabels: constLabels,
				},
				[]string{"group", "source", "split_address"},
			),
			hash: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "hash",
					Help:        "The hash of the split matches expected hash.",
					ConstLabels: constLabels,
				},
				[]string{"group", "source", "split_address", "expected_hash", "actual_hash"},
			),
			controller: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "controller",
					Help:        "The controller of the split matches expected controller.",
					ConstLabels: constLabels,
				},
				[]string{"group", "source", "split_address", "expected_controller", "actual_controller"},
			),
		}

		prometheus.MustRegister(metricsInstance.balance)
		prometheus.MustRegister(metricsInstance.hash)
		prometheus.MustRegister(metricsInstance.controller)
	})

	return metricsInstance
}

func (m Metrics) UpdateBalance(balance float64, labels []string) {
	m.balance.WithLabelValues(labels...).Set(balance)
}

func (m Metrics) UpdateHash(hash float64, labels []string) {
	m.hash.WithLabelValues(labels...).Set(hash)
}

func (m Metrics) UpdateController(controller float64, labels []string) {
	m.controller.WithLabelValues(labels...).Set(controller)
}
