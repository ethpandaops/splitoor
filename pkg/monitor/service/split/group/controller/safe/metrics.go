package safe

// transaction_queue_count
// transaction_queued shite

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	transactionQueueSize         *prometheus.GaugeVec
	transactionRecoveryNext      *prometheus.GaugeVec
	transactionRecoveryExists    *prometheus.GaugeVec
	transactionRecoveryPreSigned *prometheus.GaugeVec
	transactionRecoveryValid     *prometheus.GaugeVec
}

var (
	metricsInstance *Metrics
	once            sync.Once
)

func GetMetricsInstance(namespace, monitor string) *Metrics {
	once.Do(func() {
		constLabels := prometheus.Labels{"monitor": monitor}

		metricsInstance = &Metrics{
			transactionQueueSize: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "transaction_queue_size",
					Help:        "The number of transactions in the transaction queue.",
					ConstLabels: constLabels,
				},
				[]string{"group", "controller", "source"},
			),
			transactionRecoveryNext: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "transaction_recovery_next",
					Help:        "Whether the recovery transaction is next.",
					ConstLabels: constLabels,
				},
				[]string{"group", "controller", "source"},
			),
			transactionRecoveryExists: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "transaction_recovery_exists",
					Help:        "Whether the recovery transaction exists.",
					ConstLabels: constLabels,
				},
				[]string{"group", "controller", "source"},
			),
			transactionRecoveryPreSigned: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "transaction_recovery_pre_signed",
					Help:        "Whether the recovery transaction is pre-signed.",
					ConstLabels: constLabels,
				},
				[]string{"group", "controller", "source", "required", "submitted"},
			),
			transactionRecoveryValid: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "transaction_recovery_valid",
					Help:        "Whether the recovery transaction is valid.",
					ConstLabels: constLabels,
				},
				[]string{"group", "controller", "source"},
			),
		}

		prometheus.MustRegister(metricsInstance.transactionQueueSize)
		prometheus.MustRegister(metricsInstance.transactionRecoveryNext)
		prometheus.MustRegister(metricsInstance.transactionRecoveryExists)
		prometheus.MustRegister(metricsInstance.transactionRecoveryPreSigned)
		prometheus.MustRegister(metricsInstance.transactionRecoveryValid)
	})

	return metricsInstance
}

func (m Metrics) UpdateTransactionQueueSize(transactionQueueSize float64, labels []string) {
	m.transactionQueueSize.WithLabelValues(labels...).Set(transactionQueueSize)
}

func (m Metrics) UpdateTransactionRecoveryNext(transactionRecoveryNext float64, labels []string) {
	m.transactionRecoveryNext.WithLabelValues(labels...).Set(transactionRecoveryNext)
}

func (m Metrics) UpdateTransactionRecoveryExists(transactionRecoveryExists float64, labels []string) {
	m.transactionRecoveryExists.WithLabelValues(labels...).Set(transactionRecoveryExists)
}

func (m Metrics) UpdateTransactionRecoveryPreSigned(transactionRecoveryPreSigned float64, labels []string) {
	m.transactionRecoveryPreSigned.WithLabelValues(labels...).Set(transactionRecoveryPreSigned)
}

func (m Metrics) UpdateTransactionRecoveryValid(transactionRecoveryValid float64, labels []string) {
	m.transactionRecoveryValid.WithLabelValues(labels...).Set(transactionRecoveryValid)
}
