package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
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

func GetMetricJson(w http.ResponseWriter, r *http.Request) {
	var metric handlers.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, fmt.Sprintf("invalid json body: %v", err), http.StatusBadRequest)
		return
	}

	value, err := Storage.GetMetric(metric.MType, metric.ID)
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

func UpdateMetricJson(w http.ResponseWriter, r *http.Request) {
	var metric handlers.Metrics

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

	err := Storage.UpdateMetric(metric.MType, metric.ID, value)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid update metric %s: %v", metric.ID, err), http.StatusBadRequest)
		return
	}

	newValue, err := Storage.GetMetric(metric.MType, metric.ID)
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

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(lw, r)

		duration := time.Since(start)

		log.Info().
			Str("uri", r.RequestURI).
			Str("method", r.Method).
			Int("status", responseData.status).
			Dur("duration", duration).
			Int("size", responseData.size).
			Str("data", responseData.data).
			Msg("new request")
	}

	return http.HandlerFunc(logFn)
}
