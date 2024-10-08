package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/metrics"
	"github.com/go-chi/chi/v5"
)

var Storage collector.Collector

// type MetricCallback func(w http.ResponseWriter, r *http.Request, name, value string)

// func removeEmptyStrings(input []string) []string {
// 	var result []string
// 	for _, str := range input {
// 		if str != "" {
// 			result = append(result, str)
// 		}
// 	}
// 	return result
// }

// func MetricMiddleware(next MetricCallback) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			http.Error(w, "only POST requests are allowed!", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		if r.Header.Get("Content-Type") != "text/plain" {
// 			http.Error(w, "only Content-Type \"text/plain\" are allowed!", http.StatusBadRequest)
// 			return
// 		}

// 		segments := strings.Split(r.URL.Path, "/")
// 		segments = removeEmptyStrings(segments)

// 		if len(segments) == 2 {
// 			http.Error(w, "name metric not found", http.StatusNotFound)
// 			return
// 		}

// 		if len(segments) != 4 {
// 			http.Error(w, "invalid format request", http.StatusBadRequest)
// 			return
// 		}

// 		name := segments[2]
// 		value := segments[3]

// 		next(w, r, name, value)
// 	})
// }

// func UnknownMetricHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "only POST requests are allowed!", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	http.Error(w, "unknown type metric!", http.StatusBadRequest)
// }

func ConvertByType(metric, value string) (any, error) {
	switch metric {
	case metrics.TypeMetricCounter:
		v, err := strconv.ParseInt(value, 10, 64)
		return v, err
	case metrics.TypeMetricGauge:
		v, err := strconv.ParseFloat(value, 64)
		return v, err
	default:
		return nil, collector.ErrInvalidMetricType
	}
}

func GetListMetrics(w http.ResponseWriter, r *http.Request) {
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
	message := fmt.Sprintf("%10s (%5s): %v\n", metricName, metricType, value)
	if _, err := w.Write([]byte(message)); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
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
