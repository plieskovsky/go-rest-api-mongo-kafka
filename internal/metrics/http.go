package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	pathLabel       = "path"
	methodLabel     = "method"
	statusCodeLabel = "status"
)

var (
	once                    sync.Once
	httpRequestDurationSecs *prometheus.HistogramVec
)

// RegisterHTTPMetrics registers the HTTP prometheus metrics.
func RegisterHTTPMetrics() {
	once.Do(func() {
		httpRequestDurationSecs = promauto.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "user_service",
			Name:      "http_request_duration_seconds",
		}, []string{
			pathLabel,
			methodLabel,
			statusCodeLabel,
		})
	})
}

// HTTPRequestDurationMetricsMiddleware returns HTTP middleware that collects request duration metric.
func HTTPRequestDurationMetricsMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		start := time.Now()

		// to reduce the cardinality of metric
		path := removeDynamicPathParams(c.Request.URL.Path)

		c.Next()

		duration := time.Now().Sub(start)
		method := c.Request.Method
		statusCode := c.Writer.Status()
		CollectHTTPRequestDuration(duration, statusCode, path, method)
	}
}

// CollectHTTPRequestDuration collects the request duration metric.
func CollectHTTPRequestDuration(duration time.Duration, statusCode int, path, method string) {
	httpRequestDurationSecs.With(prometheus.Labels{
		pathLabel:       path,
		methodLabel:     method,
		statusCodeLabel: strconv.Itoa(statusCode),
	}).Observe(duration.Seconds())
}

func removeDynamicPathParams(path string) string {
	// strip the path params
	paramsSplit := strings.Split(path, "?")

	pathParts := strings.Split(paramsSplit[0], "/")
	if len(pathParts) <= 3 {
		return path
	}

	// if there are more than 2 sub paths in the path of this service API the extra ones are dynamic
	// e.g. /v1/users vs /v1/users/<id>
	return strings.Join(pathParts[0:3], "/")
}
