package main

import (
	"net/http"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/server"
)

func main() {
	storage := collector.NewMemStorage()
	server.Storage = &storage

	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, server.UnknownMetricHandler)
	mux.HandleFunc(`/update/counter/`, server.MetricMiddleware(server.CounterHandler).ServeHTTP)
	mux.HandleFunc(`/update/gauge/`, server.MetricMiddleware(server.GaugeHandler).ServeHTTP)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
