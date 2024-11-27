package serialization

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

func Load(collector collector.Collector, path string) error {
	text, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("can't read file %s: %w", path, err)
	}

	var data []Metrics

	err = json.Unmarshal(text, &data)
	if err != nil {
		return fmt.Errorf("cant parse json database: %w", err)
	}

	for _, metric := range data {
		if metric.MType == metrics.TypeMetricCounter {
			if err := collector.UpdateMetric(metric.MType, metric.ID, *metric.Delta); err != nil {
				return fmt.Errorf("update metric error: %w", err)
			}
		}

		if metric.MType == metrics.TypeMetricGauge {
			if err := collector.UpdateMetric(metric.MType, metric.ID, *metric.Value); err != nil {
				return fmt.Errorf("update metric error: %w", err)
			}
		}
	}

	return nil
}

func Save(collector collector.Collector, path string) error {
	allMetrics := collector.GetAllMetrics()

	data := make([]Metrics, 0, len(allMetrics))

	for _, metric := range allMetrics {
		var newMetric Metrics

		newMetric.ID = metric.Name()
		newMetric.MType = metric.Type()

		if newMetric.MType == metrics.TypeMetricCounter {
			val, ok := metric.Value().(int64)
			if !ok {
				return fmt.Errorf("invalid type metric")
			}

			newMetric.Delta = &val
		}

		if newMetric.MType == metrics.TypeMetricGauge {
			val, ok := metric.Value().(float64)
			if !ok {
				return fmt.Errorf("invalid type metric")
			}

			newMetric.Value = &val
		}

		data = append(data, newMetric)
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshalling: %w", err)
	}

	err = os.WriteFile(path, bytes, 0644)
	if err != nil {
		return fmt.Errorf("invalid write file %s: %w", path, err)
	}

	return nil
}
