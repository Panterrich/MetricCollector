package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
	"github.com/Panterrich/MetricCollector/pkg/serialization"
)

type CollectorHandler func(c collector.Collector, w http.ResponseWriter, r *http.Request)

func WithCollector(c collector.Collector, next CollectorHandler) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		next(c, w, r)
	}

	return http.HandlerFunc(fn)
}

func GetListMetrics(c collector.Collector, w http.ResponseWriter, r *http.Request) {
	metrics := c.GetAllMetrics(r.Context())

	w.Header().Set("Content-Type", "text/html")

	for _, metric := range metrics {
		message := fmt.Sprintf("%10s (%5s): %v\n", metric.Name(), metric.Type(), metric.Value())
		if _, err := w.Write([]byte(message)); err != nil {
			panic(err)
		}
	}
}

func GetMetric(c collector.Collector, w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	value, err := c.GetMetric(r.Context(), metricType, metricName)
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

func UpdateMetric(c collector.Collector, w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	value, err := serialization.ConvertByType(metricType, metricValue)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid value of metric: %v", err), http.StatusBadRequest)
		return
	}

	err = c.UpdateMetric(r.Context(), metricType, metricName, value)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid update metric %s: %v", metricName, err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetMetricJSON(c collector.Collector, w http.ResponseWriter, r *http.Request) {
	var metric serialization.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, fmt.Sprintf("invalid json body: %v", err), http.StatusBadRequest)
		return
	}

	value, err := c.GetMetric(r.Context(), metric.MType, metric.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("metric %s not found", metric.ID), http.StatusNotFound)
		return
	}

	if val, ok := value.(int64); ok {
		metric.Delta = &val
	} else if val, ok := value.(float64); ok {
		metric.Value = &val
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(&metric)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid json body: %v", err), http.StatusInternalServerError)
		return
	}
}

func UpdateMetricJSON(c collector.Collector, w http.ResponseWriter, r *http.Request) {
	var metric serialization.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, fmt.Sprintf("invalid json body: %v", err), http.StatusBadRequest)
		return
	}

	var value any

	if metric.MType == metrics.TypeMetricCounter {
		value = *metric.Delta
	} else {
		value = *metric.Value
	}

	err := c.UpdateMetric(r.Context(), metric.MType, metric.ID, value)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid update metric %s: %v", metric.ID, err), http.StatusBadRequest)
		return
	}

	newValue, err := c.GetMetric(r.Context(), metric.MType, metric.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("metric %s not found", metric.ID), http.StatusNotFound)
		return
	}

	if val, ok := newValue.(int64); ok {
		metric.Delta = &val
	} else if val, ok := newValue.(float64); ok {
		metric.Value = &val
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(&metric)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid json body: %v", err), http.StatusInternalServerError)
		return
	}
}
