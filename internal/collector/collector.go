package collector

import (
	"context"
	"errors"

	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

type Collector interface {
	GetMetric(ctx context.Context, kind, name string) (any, error)
	GetAllMetrics(ctx context.Context) []metrics.Metric
	UpdateMetric(ctx context.Context, kind, name string, value any) error
	UpdateMetrics(ctx context.Context, metrics []metrics.Metric) error
	Close()
}

var (
	ErrMetricNotFound    = errors.New("metric not found")
	ErrInvalidMetricType = errors.New("invalid metric type")
	ErrUpdateMetric      = errors.New("metric can't update")
)
