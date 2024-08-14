package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "status", "uri"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
}

func RecordRequest(method, status, uri string) {
	httpRequestsTotal.WithLabelValues(method, status, uri).Inc()
}

// StartMetricsServer стартует сервер для доступа к метрикам
func StartMetricsServer(port string) {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+port, nil)
}
