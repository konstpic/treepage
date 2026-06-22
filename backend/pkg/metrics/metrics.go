package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "treepage_http_requests_total",
			Help: "Total HTTP requests by service, method, path pattern and status class.",
		},
		[]string{"service", "method", "path", "status_class"},
	)
	RAGAskDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "treepage_rag_ask_duration_seconds",
			Help:    "RAG ask latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"},
	)
	SyncJobsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "treepage_sync_jobs_total",
			Help: "Git sync jobs by trigger and outcome.",
		},
		[]string{"trigger", "outcome"},
	)
	SearchRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "treepage_search_requests_total",
			Help: "Search requests by backend and outcome.",
		},
		[]string{"backend", "outcome"},
	)
)

func Register() {
	prometheus.MustRegister(HTTPRequestsTotal, RAGAskDuration, SyncJobsTotal, SearchRequestsTotal)
}

func StatusClass(code int) string {
	switch {
	case code >= 500:
		return "5xx"
	case code >= 400:
		return "4xx"
	case code >= 300:
		return "3xx"
	case code >= 200:
		return "2xx"
	default:
		return "other"
	}
}

func ObserveHTTP(service, method, path string, status int) {
	HTTPRequestsTotal.WithLabelValues(service, method, path, StatusClass(status)).Inc()
}

func ObserveRAG(duration time.Duration, err error) {
	result := "ok"
	if err != nil {
		result = "error"
	}
	RAGAskDuration.WithLabelValues(result).Observe(duration.Seconds())
}

func IncSync(trigger, outcome string) {
	SyncJobsTotal.WithLabelValues(trigger, outcome).Inc()
}

func IncSearch(backend, outcome string) {
	SearchRequestsTotal.WithLabelValues(backend, outcome).Inc()
}
