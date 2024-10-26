package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/metrics"
)

var Storage collector.Collector

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
		return nil, collector.ErrInvalidMetricType
	}
}

func GetListMetrics(w http.ResponseWriter, _ *http.Request) {
	metrics := Storage.GetAllMetrics()

	w.Header().Set("Content-Type", "text/html")

	for _, metric := range metrics {
		message := fmt.Sprintf("%10s (%5s): %v\n", metric.Name(), metric.Type(), metric.Value())
		if _, err := w.Write([]byte(message)); err != nil {
			panic(err)
		}
	}
}

func GetMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	value, err := Storage.GetMetric(metricType, metricName)
	if err != nil {
		http.Error(w, fmt.Sprintf("metric %s not found", metricName), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	message := fmt.Sprintf("%v\n", value)

	if _, err := w.Write([]byte(message)); err != nil {
		panic(err)
	}
}

func UpdateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	value, err := ConvertByType(metricType, metricValue)
	if err != nil {
		http.Error(w, "invalid value of metric", http.StatusBadRequest)
		return
	}

	err = Storage.UpdateMetric(metricType, metricName, value)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid update metric %s: %v", metricName, err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
