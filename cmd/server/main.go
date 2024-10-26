package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	memstorage "github.com/Panterrich/MetricCollector/internal/mem-storage"
	"github.com/Panterrich/MetricCollector/internal/metrics"
)

var collector memstorage.Collector

type MetricCallback func(w http.ResponseWriter, r *http.Request, name, value string)

func removeEmptyStrings(input []string) []string {
	var result []string
	for _, str := range input {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}

func MetricMiddleware(next MetricCallback) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "text/plain" {
			http.Error(w, "only Content-Type \"text/plain\" are allowed!", http.StatusBadRequest)
			return
		}

		segments := strings.Split(r.URL.Path, "/")
		segments = removeEmptyStrings(segments)

		if len(segments) == 2 {
			http.Error(w, "name metric not found", http.StatusNotFound)
			return
		}

		if len(segments) != 4 {
			http.Error(w, "invalid format request", http.StatusBadRequest)
			return
		}

		name := segments[2]
		value := segments[3]

		next(w, r, name, value)
	})
}

func unknownMetricHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	http.Error(w, "unknown type metric!", http.StatusBadRequest)
}

func counterHandler(w http.ResponseWriter, r *http.Request, name, value string) {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		http.Error(w, "invalid value of metric", http.StatusBadRequest)
		return
	}

	err = collector.UpdateMetric(metrics.TypeMetricCounter, name, v)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid update metric %s: %v", name, err)))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func gaugeHandler(w http.ResponseWriter, r *http.Request, name, value string) {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		http.Error(w, "invalid value of metric", http.StatusBadRequest)
		return
	}

	err = collector.UpdateMetric(metrics.TypeMetricGauge, name, v)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid update metric %s: %v", name, err)))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	storage := memstorage.NewMemStorage()
	collector = &storage

	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, unknownMetricHandler)
	mux.HandleFunc(`/update/counter/`, MetricMiddleware(counterHandler).ServeHTTP)
	mux.HandleFunc(`/update/gauge/`, MetricMiddleware(gaugeHandler).ServeHTTP)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
