package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lb_requests_total",
			Help: "Total number of requests handled by the load balancer.",
		},
		[]string{"method", "path", "status", "backend"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "lb_request_duration_seconds",
			Help:    "Request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	ProxyErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lb_proxy_errors_total",
			Help: "Total number of proxy errors by backend.",
		},
		[]string{"backend"},
	)

	BackendHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lb_backend_health",
			Help: "Backend health status: 1 = healthy, 0 = unhealthy.",
		},
		[]string{"backend"},
	)
)

func Register() {
	prometheus.MustRegister(
		RequestsTotal,
		RequestDuration,
		ProxyErrorsTotal,
		BackendHealth,
	)
}
