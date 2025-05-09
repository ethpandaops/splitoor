package eoa

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	balance *prometheus.GaugeVec
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
					Help:        "The balance of the eoa.",
					ConstLabels: constLabels,
				},
				[]string{"group", "controller", "source", "address"},
			),
		}

		prometheus.MustRegister(metricsInstance.balance)
	})

	return metricsInstance
}

func (m Metrics) UpdateBalance(balance float64, labels []string) {
	m.balance.WithLabelValues(labels...).Set(balance)
}
