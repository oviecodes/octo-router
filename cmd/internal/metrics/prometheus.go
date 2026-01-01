package metrics

import (
	"llm-router/utils"
	"net/http"
	"os"

	// "time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	// "github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_router_http_requests_total",
			Help: "Total HTTP requests received",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_router_http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	ProviderRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_router_provider_requests_total",
			Help: "Total requests to each provider",
		},
		[]string{"provider", "status"},
	)

	ProviderRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_router_provider_request_duration_seconds",
			Help:    "Provider request latency",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"provider"},
	)

	ProviderTokensUsed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_router_provider_tokens_total",
			Help: "Total tokens consumed per provider",
		},
		[]string{"provider", "type"},
	)

	ProviderCostTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_router_provider_cost_usd_total",
			Help: "Total cost in USD per provider",
		},
		[]string{"provider"},
	)

	CircuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "llm_router_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half_open)",
		},
		[]string{"provider"},
	)

	CircuitBreakerTrips = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_router_circuit_breaker_trips_total",
			Help: "Number of times circuit breaker opened",
		},
		[]string{"provider"},
	)

	RetryAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_router_retry_attempts_total",
			Help: "Total retry attempts",
		},
		[]string{"provider", "outcome"},
	)

	CacheHitsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "llm_router_cache_hits_total",
			Help: "Total cache hits",
		},
	)

	CacheMissesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "llm_router_cache_misses_total",
			Help: "Total cache misses",
		},
	)
)

func Metrics() error {

	var logger = utils.SetUpLogger()
	mux := http.NewServeMux()
	reg := prometheus.NewRegistry()

	reg.MustRegister(
		HttpRequestsTotal,
		HttpRequestDuration,
		ProviderCostTotal,
		ProviderRequestDuration,
		ProviderRequestsTotal,
		ProviderTokensUsed,
		CacheHitsTotal,
		CacheMissesTotal,
		CircuitBreakerState,
		CircuitBreakerTrips,
		RetryAttemptsTotal,
	)

	mux.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{},
	))

	port := os.Getenv("METRICS_PORT")
	if port == "" {
		port = "9090"
	}

	logger.Info("Starting Metrics server on port", zap.String("port", port))
	return http.ListenAndServe(":"+port, mux)
}
