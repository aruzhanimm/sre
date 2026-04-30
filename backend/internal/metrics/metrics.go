package metrics

import (
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusMetrics struct {
    requestCounter  *prometheus.CounterVec
    requestDuration *prometheus.HistogramVec
}

func NewPrometheusMetrics(db *pgxpool.Pool) (*PrometheusMetrics, *prometheus.Registry) {
    registry := prometheus.NewRegistry()

    requestCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "betkz_http_requests_total",
        Help: "Total number of HTTP requests processed by BetKZ.",
    }, []string{"method", "path", "status"})

    requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "betkz_http_request_duration_seconds",
        Help:    "HTTP request duration in seconds.",
        Buckets: prometheus.DefBuckets,
    }, []string{"method", "path", "status"})

    registry.MustRegister(
        prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
        prometheus.NewGoCollector(),
        requestCounter,
        requestDuration,
        prometheus.NewGaugeFunc(prometheus.GaugeOpts{
            Name: "betkz_db_total_connections",
            Help: "Total PostgreSQL connection count.",
        }, func() float64 { return float64(db.Stat().TotalConns()) }),
        prometheus.NewGaugeFunc(prometheus.GaugeOpts{
            Name: "betkz_db_acquired_connections",
            Help: "Currently acquired PostgreSQL connections.",
        }, func() float64 { return float64(db.Stat().AcquiredConns()) }),
    )

    return &PrometheusMetrics{
        requestCounter:  requestCounter,
        requestDuration: requestDuration,
    }, registry
}

func (m *PrometheusMetrics) GinMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()

        statusCode := strconv.Itoa(c.Writer.Status())
        path := c.FullPath()
        if path == "" {
            path = c.Request.URL.Path
        }

        m.requestCounter.WithLabelValues(c.Request.Method, path, statusCode).Inc()
        m.requestDuration.WithLabelValues(c.Request.Method, path, statusCode).Observe(time.Since(start).Seconds())
    }
}

func MetricsHandler(registry *prometheus.Registry) http.Handler {
    return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}
