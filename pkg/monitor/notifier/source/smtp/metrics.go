package smtp

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	published *prometheus.CounterVec
	errors    *prometheus.CounterVec
}

var (
	metricsInstance *Metrics
	once            sync.Once
)

func GetMetricsInstance(namespace, monitor string) *Metrics {
	once.Do(func() {
		constLabels := prometheus.Labels{"monitor": monitor}
		labels := []string{"group", "source", "source_type"}
		errorLabels := []string{"group", "source", "source_type", "error_type"}

		metricsInstance = &Metrics{
			published: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Namespace:   namespace,
					Name:        "published_total",
					Help:        "Total number of messages published via smtp",
					ConstLabels: constLabels,
				},
				labels,
			),
			errors: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Namespace:   namespace,
					Name:        "errors_total",
					Help:        "Total number of smtp errors by type",
					ConstLabels: constLabels,
				},
				errorLabels,
			),
		}

		prometheus.MustRegister(metricsInstance.published)
		prometheus.MustRegister(metricsInstance.errors)
	})

	return metricsInstance
}

func (m Metrics) IncMessagesPublished(group string, source string, sourceType string) {
	m.published.WithLabelValues(group, source, sourceType).Inc()
}

func (m Metrics) IncErrors(group string, source string, sourceType string, errorType string) {
	m.errors.WithLabelValues(group, source, sourceType, errorType).Inc()
}
