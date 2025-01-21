package safe

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	requests        *prometheus.CounterVec
	responses       *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewMetrics(namespace, monitorName string) Metrics {
	constLabels := prometheus.Labels{"monitor": monitorName}

	m := Metrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Name:        "request_count",
			Help:        "Number of requests",
			ConstLabels: constLabels,
		}, []string{"method", "endpoint", "path", "chain_id", "safe_address"}),
		responses: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Name:        "response_count",
			Help:        "Number of responses",
			ConstLabels: constLabels,
		}, []string{"method", "endpoint", "path", "code", "chain_id", "safe_address"}),
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   namespace,
			Name:        "request_duration_seconds",
			Help:        "Request duration (in seconds.)",
			Buckets:     []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			ConstLabels: constLabels,
		}, []string{"method", "endpoint", "path", "code", "chain_id", "safe_address"}),
	}

	prometheus.MustRegister(m.requests)
	prometheus.MustRegister(m.responses)
	prometheus.MustRegister(m.requestDuration)

	return m
}

func (m Metrics) ObserveRequest(method, endpoint, path, chainID, safeAddress string) {
	m.requests.WithLabelValues(method, endpoint, path, chainID, safeAddress).Inc()
}

func (m Metrics) ObserveResponse(method, endpoint, path, code, chainID, safeAddress string, duration time.Duration) {
	m.responses.WithLabelValues(method, endpoint, path, code, chainID, safeAddress).Inc()
	m.requestDuration.WithLabelValues(method, endpoint, path, code, chainID, safeAddress).Observe(duration.Seconds())
}
