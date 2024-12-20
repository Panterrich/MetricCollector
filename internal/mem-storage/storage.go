package memstorage

import (
	"sync"

	"github.com/Panterrich/MetricCollector/internal/metrics"
)

type MemStorage struct {
	mutex   sync.RWMutex
	storage map[string]map[string]metrics.Metric
}

var _ Collector = (*MemStorage)(nil)

func NewMemStorage() MemStorage {
	return MemStorage{
		storage: make(map[string]map[string]metrics.Metric),
	}
}

func (m *MemStorage) GetMetric(kind, name string) (any, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	specificMetrics, ok := m.storage[kind]
	if !ok {
		return nil, ErrInvalidMetricType
	}

	metric, ok := specificMetrics[name]
	if !ok {
		return nil, ErrMetricNotFound
	}

	return metric.Value(), nil
}

func (m *MemStorage) UpdateMetric(kind, name string, value any) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	specificMetrics, ok := m.storage[kind]
	if !ok {
		specificMetrics = make(map[string]metrics.Metric)
		m.storage[kind] = specificMetrics
	}

	metric, ok := specificMetrics[name]
	if !ok {
		metric = metrics.NewMetric(kind, name)
		if metric == nil {
			return ErrInvalidMetricType
		}
		specificMetrics[name] = metric
	}

	if !metric.Update(value) {
		return ErrUpdateMetric
	}

	return nil
}
