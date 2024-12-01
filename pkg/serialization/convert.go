package serialization

import (
	"fmt"
	"strconv"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

func ConvertByType(metric, value string) (any, error) {
	switch metric {
	case metrics.TypeMetricCounter:
		if v, err := strconv.ParseInt(value, 10, 64); err != nil {
			return nil, fmt.Errorf("invalid parse int: %w", err)
		} else {
			return v, nil
		}
	case metrics.TypeMetricGauge:
		if v, err := strconv.ParseFloat(value, 64); err != nil {
			return nil, fmt.Errorf("invalid parse float: %w", err)
		} else {
			return v, nil
		}
	default:
		return nil, fmt.Errorf("%w: %s", collector.ErrInvalidMetricType, metric)
	}
}

func ConvertToMetrics(jsonMetrics []Metrics) ([]metrics.Metric, error) {
	m := make([]metrics.Metric, 0, len(jsonMetrics))

	for _, jsonMetric := range jsonMetrics {
		switch jsonMetric.MType {
		case metrics.TypeMetricCounter:
			if jsonMetric.Delta == nil {
				return nil, collector.ErrMetricNotFound
			}

			metric := metrics.NewCounter(jsonMetric.ID)
			metric.Update(*jsonMetric.Delta)
			m = append(m, metric)

		case metrics.TypeMetricGauge:
			if jsonMetric.Value == nil {
				return nil, collector.ErrMetricNotFound
			}

			metric := metrics.NewGauge(jsonMetric.ID)
			metric.Update(*jsonMetric.Value)
			m = append(m, metric)

		default:
			return nil, collector.ErrInvalidMetricType
		}
	}

	return m, nil
}

func ConvertToJSONMetrics(m []metrics.Metric) ([]Metrics, error) {
	jsonMetrics := make([]Metrics, 0, len(m))

	for _, metric := range m {
		jsonMetric := Metrics{
			ID:    metric.Name(),
			MType: metric.Type(),
		}

		err := jsonMetric.SetValue(metric.Value())
		if err != nil {
			return nil, fmt.Errorf("json metric set value: %w", err)
		}

		jsonMetrics = append(jsonMetrics, jsonMetric)
	}

	return jsonMetrics, nil
}
