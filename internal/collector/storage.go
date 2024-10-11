package collector

import (
	"sync"

	"github.com/Panterrich/MetricCollector/internal/metrics"
)

type MemStorage struct {
	lock    sync.RWMutex
	storage map[string]map[string]metrics.Metric
}

var _ Collector = (*MemStorage)(nil)

func NewMemStorage() MemStorage {
	return MemStorage{
		lock:    sync.RWMutex{},
		storage: make(map[string]map[string]metrics.Metric),
	}
}

func (m *MemStorage) GetMetric(kind, name string) (any, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

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

func (m *MemStorage) GetAllMetrics() []metrics.Metric {
	var res []metrics.Metric

	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, specMetrics := range m.storage {
		for _, metric := range specMetrics {
			res = append(res, metric)
		}
	}

	return res
}

func (m *MemStorage) UpdateMetric(kind, name string, value any) error {
	m.lock.Lock()
	defer m.lock.Unlock()

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
