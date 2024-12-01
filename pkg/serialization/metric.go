package serialization

import (
	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (m *Metrics) GetValue() (any, error) {
	switch m.MType {
	case metrics.TypeMetricCounter:
		if m.Delta == nil {
			return nil, collector.ErrMetricNotFound
		}

		return *m.Delta, nil

	case metrics.TypeMetricGauge:
		if m.Value == nil {
			return nil, collector.ErrMetricNotFound
		}

		return *m.Value, nil

	default:
		return nil, collector.ErrInvalidMetricType
	}
}

func (m *Metrics) SetValue(value any) error {
	switch m.MType {
	case metrics.TypeMetricCounter:
		val, ok := value.(int64)
		if !ok {
			return collector.ErrUpdateMetric
		}

		m.Delta = &val

		return nil
	case metrics.TypeMetricGauge:
		val, ok := value.(float64)
		if !ok {
			return collector.ErrUpdateMetric
		}

		m.Value = &val

		return nil
	default:
		return collector.ErrInvalidMetricType
	}
}
