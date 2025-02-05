package group

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	balance      *prometheus.GaugeVec
	hashStable   *prometheus.GaugeVec
	hashInitial  *prometheus.GaugeVec
	hashRecovery *prometheus.GaugeVec
	controller   *prometheus.GaugeVec
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
			hashStable: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "hash_stable",
					Help:        "The hash of the split matches expected stable hash.",
					ConstLabels: constLabels,
				},
				[]string{"group", "source", "split_address", "expected_hash", "actual_hash"},
			),
			hashInitial: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "hash_initial",
					Help:        "The hash of the split matches expected initial hash.",
					ConstLabels: constLabels,
				},
				[]string{"group", "source", "split_address", "expected_hash", "actual_hash"},
			),
			hashRecovery: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "hash_recovery",
					Help:        "The hash of the split matches expected recovery hash.",
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
				[]string{"group", "source", "split_address", "expected_controller", "actual_controller", "type"},
			),
		}

		prometheus.MustRegister(metricsInstance.balance)
		prometheus.MustRegister(metricsInstance.hashStable)
		prometheus.MustRegister(metricsInstance.hashInitial)
		prometheus.MustRegister(metricsInstance.hashRecovery)
		prometheus.MustRegister(metricsInstance.controller)
	})

	return metricsInstance
}

func (m Metrics) UpdateBalance(balance float64, labels []string) {
	m.balance.WithLabelValues(labels...).Set(balance)
}

func (m Metrics) UpdateHashStable(hash float64, labels []string) {
	m.hashStable.WithLabelValues(labels...).Set(hash)
}

func (m Metrics) UpdateHashInitial(hash float64, labels []string) {
	m.hashInitial.WithLabelValues(labels...).Set(hash)
}

func (m Metrics) UpdateHashRecovery(hash float64, labels []string) {
	m.hashRecovery.WithLabelValues(labels...).Set(hash)
}

func (m Metrics) UpdateController(controller float64, labels []string) {
	m.controller.WithLabelValues(labels...).Set(controller)
}
