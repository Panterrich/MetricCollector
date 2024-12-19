package storages

import (
	"context"
	"fmt"
	"sync"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

type element struct {
	sync.RWMutex
	metrics.Metric
}

type Memory struct {
	lock    sync.RWMutex
	storage map[string]map[string]*element
}

var _ collector.Collector = (*Memory)(nil)

func NewMemory() collector.Collector {
	return &Memory{
		lock:    sync.RWMutex{},
		storage: make(map[string]map[string]*element),
	}
}

func (m *Memory) GetMetric(_ context.Context, kind, name string) (any, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	specificMetrics, ok := m.storage[kind]
	if !ok {
		return nil, collector.ErrInvalidMetricType
	}

	metric, ok := specificMetrics[name]
	if !ok {
		return nil, collector.ErrMetricNotFound
	}

	metric.RLock()
	defer metric.RUnlock()

	return metric.Value(), nil
}

func (m *Memory) GetAllMetrics(_ context.Context) []metrics.Metric {
	var res []metrics.Metric

	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, specMetrics := range m.storage {
		for _, metric := range specMetrics {
			metric.RLock()
			res = append(res, metrics.Clone(metric.Metric))
			metric.RUnlock()
		}
	}

	return res
}

func (m *Memory) UpdateMetric(_ context.Context, kind, name string, value any) error {
	// fast path
	m.lock.RLock()
	specificMetrics, hasKind := m.storage[kind]

	if hasKind {
		metric, ok := specificMetrics[name]
		if ok {
			defer m.lock.RUnlock()

			metric.Lock()
			defer metric.Unlock()

			if !metric.Update(value) {
				return fmt.Errorf("%s(%s): %w", name, kind, collector.ErrUpdateMetric)
			}

			return nil
		}
	}
	m.lock.RUnlock()

	m.lock.Lock()
	defer m.lock.Unlock()

	specificMetrics, hasKind = m.storage[kind]

	if !hasKind {
		specificMetrics = make(map[string]*element)
		m.storage[kind] = specificMetrics
	}

	metric, ok := specificMetrics[name]

	if !ok {
		metric = &element{Metric: metrics.NewMetric(kind, name)}
		if metric.Metric == nil {
			return collector.ErrInvalidMetricType
		}

		specificMetrics[name] = metric
	}

	if !metric.Update(value) {
		return fmt.Errorf("%s(%s): %w", name, kind, collector.ErrUpdateMetric)
	}

	return nil
}

func (m *Memory) UpdateMetrics(ctx context.Context, metrics []metrics.Metric) error {
	for _, metric := range metrics {
		if err := m.UpdateMetric(ctx, metric.Type(), metric.Name(), metric.Value()); err != nil {
			return fmt.Errorf("file storage update metric: %w", err)
		}

		if ctx.Err() != nil {
			return nil
		}
	}

	return nil
}

func (m *Memory) Close() {
}
