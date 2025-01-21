package account

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	balance      *prometheus.GaugeVec
	splitBalance *prometheus.GaugeVec
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
					Help:        "The balance of the account.",
					ConstLabels: constLabels,
				},
				[]string{"group", "split", "source"},
			),
			splitBalance: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "split_balance",
					Help:        "The balance of the account on the splits contract (not yet withdrawn).",
					ConstLabels: constLabels,
				},
				[]string{"group", "split", "source", "expected_hash"},
			),
		}

		prometheus.MustRegister(metricsInstance.balance)
		prometheus.MustRegister(metricsInstance.splitBalance)
	})

	return metricsInstance
}

func (m Metrics) UpdateBalance(balance float64, labels []string) {
	m.balance.WithLabelValues(labels...).Set(balance)
}

func (m Metrics) UpdateSplitBalance(splitBalance float64, labels []string) {
	m.splitBalance.WithLabelValues(labels...).Set(splitBalance)
}
