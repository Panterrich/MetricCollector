package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/serialization"
)

type CollectorHandler func(c collector.Collector, w http.ResponseWriter, r *http.Request)

func WithCollector(c collector.Collector, next CollectorHandler) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		next(c, w, r)
	}

	return http.HandlerFunc(fn)
}

// @Summary Returns a list of metrics
// @Description Retrieves all metrics with their names, types, and values
// @Tags Metrics
// @Produce text/html
// @Success 200 {string} string "List of metrics"
// @Failure 500 {string} error "Internal server error"
// @Router / [get]
func GetListMetrics(c collector.Collector, w http.ResponseWriter, r *http.Request) {
	metrics := c.GetAllMetrics(r.Context())

	w.Header().Set("Content-Type", "text/html")

	for _, metric := range metrics {
		message := fmt.Sprintf("%10s (%5s): %v\n", metric.Name(), metric.Type(), metric.Value())
		if _, err := w.Write([]byte(message)); err != nil {
			http.Error(w, fmt.Sprintf("body writer: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

// @Summary Returns the value of a specific metric
// @Description Retrieves the value of a metric by its name and type
// @Tags Metrics
// @Produce text/plain
// @Param metricType path string true "Type of the metric (e.g., gauge, counter)"
// @Param metricName path string true "Name of the metric to be getting."
// @Success 200 {string} string "Metric Value"
// @Failure 404 {string} error "Metric Not Found"
// @Failure 500 {string} error "Internal Server Error"
// @Router /value/{metricType}/{metricName} [get]
func GetMetric(c collector.Collector, w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	value, err := c.GetMetric(r.Context(), metricType, metricName)
	if err != nil {
		http.Error(w, fmt.Sprintf("metric %s(%s) not found", metricName, metricType), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	message := fmt.Sprintf("%v\n", value)

	if _, err := w.Write([]byte(message)); err != nil {
		http.Error(w, fmt.Sprintf("body writer: %v", err), http.StatusInternalServerError)
		return
	}
}

// @Summary Updates a metric
// @Description Updates the value of an existing metric or create a new metric by its name and type
// @Tags Metrics
// @Param metricType path string true "Type of the metric (e.g., gauge, counter)"
// @Param metricName path string true "Name of the metric to be updated."
// @Param metricValue path string true "New value for the metric."
// @Success 200 "Metric updated successfully"
// @Failure 400 {string} error "Invalid input or update failure"
// @Router /update/{metricType}/{metricName}/{metricValue} [post]
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

// @Summary Returns a metric as JSON
// @Description Retrieves a metric by its name and type and returns it in JSON format
// @Tags Metrics
// @Accept application/json
// @Produce application/json
// @Success 200 {object} serialization.Metric "Getting metric details"
// @Failure 400 {string} error "Invalid JSON body"
// @Failure 404 {string} error "Metric not found"
// @Failure 500 {string} error "Internal server error"
// @Router /value/ [post]
func GetMetricJSON(c collector.Collector, w http.ResponseWriter, r *http.Request) {
	var metric serialization.Metric

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, fmt.Sprintf("invalid json body: %v", err), http.StatusBadRequest)
		return
	}

	value, err := c.GetMetric(r.Context(), metric.MType, metric.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("metric %s not found", metric.ID), http.StatusNotFound)
		return
	}

	if err = metric.SetValue(value); err != nil {
		http.Error(w, fmt.Sprintf("metric set value %s: %v", value, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(&metric)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid json body: %v", err), http.StatusInternalServerError)
		return
	}
}

// @Summary Updates a metric using JSON
// @Description Updates the value of a metric based on the provided JSON payload
// @Tags Metrics
// @Accept application/json
// @Produce application/json
// @Success 200 {object} serialization.Metric "Updated metric details"
// @Failure 400 {string} error "Invalid JSON body or invalid metric value"
// @Failure 404 {string} error "Metric not found"
// @Failure 500 {string} error "Internal server error"
// @Router /update [post]
func UpdateMetricJSON(c collector.Collector, w http.ResponseWriter, r *http.Request) {
	var metric serialization.Metric

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, fmt.Sprintf("invalid json body: %v", err), http.StatusBadRequest)
		return
	}

	value, err := metric.GetValue()
	if err != nil {
		http.Error(w, fmt.Sprintf("metric get value: %v", err), http.StatusBadRequest)
		return
	}

	err = c.UpdateMetric(r.Context(), metric.MType, metric.ID, value)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid update metric %s: %v", metric.ID, err), http.StatusBadRequest)
		return
	}

	newValue, err := c.GetMetric(r.Context(), metric.MType, metric.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("metric %s not found", metric.ID), http.StatusNotFound)
		return
	}

	err = metric.SetValue(newValue)
	if err != nil {
		http.Error(w, fmt.Sprintf("metric %s not found", metric.ID), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(&metric)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid json body: %v", err), http.StatusInternalServerError)
		return
	}
}

// @Summary Updates multiple metrics using JSON
// @Description Updates the values of multiple metrics based on the provided JSON payload
// @Tags Metrics
// @Accept application/json
// @Produce application/json
// @Success 200 "Successfully updated metrics"
// @Failure 400 {string} error "Invalid JSON body or invalid metrics"
// @Failure 500 {string} error "Internal server error"
// @Router /updates [post]
func UpdateMetricsJSON(c collector.Collector, w http.ResponseWriter, r *http.Request) {
	var jsonMetrics []serialization.Metric

	if err := json.NewDecoder(r.Body).Decode(&jsonMetrics); err != nil {
		http.Error(w, fmt.Sprintf("invalid json body: %v", err), http.StatusBadRequest)
		return
	}

	metrics, err := serialization.ConvertToMetrics(jsonMetrics)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid convert metrics: %v", err), http.StatusBadRequest)
		return
	}

	err = c.UpdateMetrics(r.Context(), metrics)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid update metrics: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
