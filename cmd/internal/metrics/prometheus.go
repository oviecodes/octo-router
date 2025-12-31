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

var HttpRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "http_requests_total",
	Help: "Total number of HTTP requests received",
}, []string{"status", "path", "method"})

func Metrics() {

	var logger = utils.SetUpLogger()
	mux := http.NewServeMux()
	reg := prometheus.NewRegistry()

	reg.MustRegister(HttpRequestCounter)

	mux.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{},
	))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	logger.Info("Starting Metrics server on port", zap.String("port", port))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatal("metrics server failed to start:", zap.Error(err))
	}
}
