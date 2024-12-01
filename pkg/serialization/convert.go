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

func ConvertToMetric(jsonMetric Metric) (metrics.Metric, error) {
	var metric metrics.Metric

	switch jsonMetric.MType {
	case metrics.TypeMetricCounter:
		metric = metrics.NewCounter(jsonMetric.ID)

		if jsonMetric.Delta != nil {
			metric.Update(*jsonMetric.Delta)
		}

	case metrics.TypeMetricGauge:
		metric = metrics.NewGauge(jsonMetric.ID)

		if jsonMetric.Val != nil {
			metric.Update(*jsonMetric.Val)
		}

	default:
		return nil, collector.ErrInvalidMetricType
	}

	return metric, nil
}

func ConvertToMetrics(jsonMetrics []Metric) ([]metrics.Metric, error) {
	m := make([]metrics.Metric, 0, len(jsonMetrics))

	for _, jsonMetric := range jsonMetrics {
		metric, err := ConvertToMetric(jsonMetric)
		if err != nil {
			return nil, fmt.Errorf("convert to metric: %w", err)
		}

		m = append(m, metric)
	}

	return m, nil
}

func ConvertToJSONMetric(m metrics.Metric) (Metric, error) {
	jsonMetric := Metric{
		ID:    m.Name(),
		MType: m.Type(),
	}

	err := jsonMetric.SetValue(m.Value())
	if err != nil {
		return Metric{}, fmt.Errorf("json metric set value: %w", err)
	}

	return jsonMetric, nil
}

func ConvertToJSONMetrics(m []metrics.Metric) ([]Metric, error) {
	jsonMetrics := make([]Metric, 0, len(m))

	for _, metric := range m {
		jsonMetric, err := ConvertToJSONMetric(metric)
		if err != nil {
			return nil, fmt.Errorf("convert to json metric: %w", err)
		}

		jsonMetrics = append(jsonMetrics, jsonMetric)
	}

	return jsonMetrics, nil
}
