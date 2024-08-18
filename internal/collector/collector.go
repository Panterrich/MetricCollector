package collector

import (
	"errors"

	"github.com/Panterrich/MetricCollector/internal/metrics"
)

type Collector interface {
	GetMetric(kind, name string) (any, error)
	GetAllMetrics() []metrics.Metric
	UpdateMetric(kind, name string, value any) error
}

var (
	ErrMetricNotFound    = errors.New("metric not found")
	ErrInvalidMetricType = errors.New("invalid metric type")
	ErrUpdateMetric      = errors.New("metric can't update")
)
