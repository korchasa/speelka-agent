// Package contracts defines the interfaces for the MCP server components.
package types

import (
	"context"
	"net/http"
	"time"
)

// MetricType represents the type of metric.
type MetricType string

const (
	// MetricTypeCounter is a metric that only increases.
	MetricTypeCounter MetricType = "COUNTER"

	// MetricTypeGauge is a metric that can increase or decrease.
	MetricTypeGauge MetricType = "GAUGE"

	// MetricTypeHistogram is a metric that samples observations and counts them in configurable buckets.
	MetricTypeHistogram MetricType = "HISTOGRAM"
)

// Metric represents a single metric.
type Metric struct {
	// Name is the name of the metric.
	Name string `json:"name"`

	// Type is the type of the metric.
	Type MetricType `json:"type"`

	// Value is the current value of the metric.
	Value interface{} `json:"value"`

	// Description provides additional information about the metric.
	Description string `json:"description,omitempty"`

	// Labels contains additional labels for the metric.
	Labels map[string]string `json:"labels,omitempty"`
}

// MetricsResponse represents the response from the metrics endpoint.
type MetricsResponse struct {
	// Metrics contains all metrics.
	Metrics []Metric `json:"metrics"`

	// Timestamp is the time when the metrics were collected.
	Timestamp time.Time `json:"timestamp"`
}

// MetricsCollectorSpec represents the interface for the metrics collector component.
type MetricsCollectorSpec interface {
	// GetMetrics returns all collected metrics.
	GetMetrics(ctx context.Context) MetricsResponse

	// IncrementCounter increments a counter metric by the given value.
	IncrementCounter(name string, value float64, labels map[string]string)

	// SetGauge sets a gauge metric to the given value.
	SetGauge(name string, value float64, labels map[string]string)

	// ObserveHistogram adds an observation to a histogram metric.
	ObserveHistogram(name string, value float64, labels map[string]string)

	// RegisterMetric registers a new metric with the collector.
	RegisterMetric(name string, metricType MetricType, description string)

	// Handler returns the HTTP handler for metrics.
	Handler() http.Handler
}
