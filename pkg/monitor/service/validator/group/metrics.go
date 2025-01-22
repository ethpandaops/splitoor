package group

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	balance             *prometheus.GaugeVec
	credentialsCode     *prometheus.GaugeVec
	lastAttestationSlot *prometheus.GaugeVec
	totalWithdrawals    *prometheus.GaugeVec
	status              *prometheus.GaugeVec
}

var (
	metricsInstance *Metrics
	once            sync.Once
)

func GetMetricsInstance(namespace, monitor string) *Metrics {
	once.Do(func() {
		constLabels := prometheus.Labels{"monitor": monitor}
		labels := []string{"group", "pubkey", "source"}

		metricsInstance = &Metrics{
			balance: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "balance",
					Help:        "The balance of the validator.",
					ConstLabels: constLabels,
				},
				labels,
			),
			credentialsCode: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "credentials_code",
					Help:        "The withdrawal credentials code of the validator.",
					ConstLabels: constLabels,
				},
				labels,
			),
			status: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "status_code",
					Help:        "The status code of the validator (0=unknown, 1=mempool, 2=deposited, 3=pending, 4=deposit_invalid, 5=active_online, 6=active_offline, 7=exiting_online, 8=exiting_offline, 9=slashing_online, 10=slashing_offline, 11=exited_unslashed, 12=exited_slashed).",
					ConstLabels: constLabels,
				},
				labels,
			),
			lastAttestationSlot: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "last_attestation_slot",
					Help:        "The last attestation slot of the validator, beaconcha.in API only.",
					ConstLabels: constLabels,
				},
				labels,
			),
			totalWithdrawals: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Name:        "total_withdrawals",
					Help:        "The total withdrawals of the validator, beaconcha.in API only",
					ConstLabels: constLabels,
				},
				labels,
			),
		}

		prometheus.MustRegister(metricsInstance.balance)
		prometheus.MustRegister(metricsInstance.credentialsCode)
		prometheus.MustRegister(metricsInstance.lastAttestationSlot)
		prometheus.MustRegister(metricsInstance.totalWithdrawals)
		prometheus.MustRegister(metricsInstance.status)
	})

	return metricsInstance
}

func (m Metrics) UpdateBalance(balance float64, labels []string) {
	m.balance.WithLabelValues(labels...).Set(balance)
}

func (m Metrics) UpdateCredentialsCode(code float64, labels []string) {
	m.credentialsCode.WithLabelValues(labels...).Set(code)
}

func (m Metrics) UpdateLastAttestationSlot(slot float64, labels []string) {
	m.lastAttestationSlot.WithLabelValues(labels...).Set(slot)
}

func (m Metrics) UpdateTotalWithdrawals(total float64, labels []string) {
	m.totalWithdrawals.WithLabelValues(labels...).Set(total)
}

func (m Metrics) UpdateStatus(status MetricsStatus, labels []string) {
	var statusCode float64

	switch status {
	case MetricsStatusMempool:
		statusCode = 1
	case MetricsStatusDeposited:
		statusCode = 2
	case MetricsStatusPending:
		statusCode = 3
	case MetricsStatusDepositInvalid:
		statusCode = 4
	case MetricsStatusActiveOnline:
		statusCode = 5
	case MetricsStatusActiveOffline:
		statusCode = 6
	case MetricsStatusExitingOnline:
		statusCode = 7
	case MetricsStatusExitingOffline:
		statusCode = 8
	case MetricsStatusSlashingOnline:
		statusCode = 9
	case MetricsStatusSlashingOffline:
		statusCode = 10
	case MetricsStatusExitedUnslashed:
		statusCode = 11
	case MetricsStatusExitedSlashed:
		statusCode = 12
	default:
		statusCode = 0
	}

	m.status.WithLabelValues(labels...).Set(statusCode)
}
