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

	specific_metrics, ok := m.storage[kind]
	if !ok {
		return nil, ErrInvalidMetricType
	}

	metric, ok := specific_metrics[name]
	if !ok {
		return nil, ErrMetricNotFound
	}

	return metric.Value(), nil
}

func (m *MemStorage) UpdateMetric(kind, name string, value any) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	specific_metrics, ok := m.storage[kind]
	if !ok {
		specific_metrics = make(map[string]metrics.Metric)
		m.storage[kind] = specific_metrics
	}

	metric, ok := specific_metrics[name]
	if !ok {
		metric = metrics.NewMetric(kind, name)
		if metric == nil {
			return ErrInvalidMetricType
		}
		specific_metrics[name] = metric
	}

	if !metric.Update(value) {
		return ErrUpdateMetric
	}

	return nil
}
