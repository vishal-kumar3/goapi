package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

// <namespace>_<subsystem>_<metric>_<unit>
// namespace: your app or service name (e.g. goapi, auth, chatserver)
// subsystem: logical grouping (e.g. http, db, cache)
// metric: what you’re measuring (e.g. requests, latency)
// unit: time/bytes/counts/etc.

var (
	namespace = "goapi"

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests processed.",
		},
		[]string{"route", "method", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "Histogram of HTTP request duration",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"route", "method"},
	)
)

func init() {
	// prometheus.MustRegister(
	// Go runtime metrics
	// collectors.NewGoCollector(),
	// collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	// )

	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
	)
}

func MetricMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		route := "unknown"
		if match := mux.CurrentRoute(r); match != nil {
			if name := match.GetName(); name != "" {
				route = name
			}
		}

		sw := &StatusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		duration := time.Since(start).Seconds()
		status := statusClass(sw.status)

		httpRequestsTotal.WithLabelValues(route, r.Method, status).Inc()
		httpRequestDuration.WithLabelValues(route, r.Method).Observe(duration)
	})
}

type StatusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *StatusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func statusClass(code int) string {
	switch {
	case code >= 100 && code < 200:
		return "1xx"
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	default:
		return "5xx"
	}
}
