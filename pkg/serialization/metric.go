package serialization

import (
	"encoding/json"
	"fmt"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Val   *float64 `json:"value,omitempty"`
}

func (m *Metric) GetValue() (any, error) {
	switch m.MType {
	case metrics.TypeMetricCounter:
		if m.Delta == nil {
			return nil, collector.ErrMetricNotFound
		}

		return *m.Delta, nil

	case metrics.TypeMetricGauge:
		if m.Val == nil {
			return nil, collector.ErrMetricNotFound
		}

		return *m.Val, nil

	default:
		return nil, collector.ErrInvalidMetricType
	}
}

func (m *Metric) SetValue(value any) error {
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

		m.Val = &val

		return nil
	default:
		return collector.ErrInvalidMetricType
	}
}

func MetricsToJSON(metrics []Metric) ([]byte, error) {
	data, err := json.MarshalIndent(&metrics, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("metrics to json: %w", err)
	}

	return data, nil
}
