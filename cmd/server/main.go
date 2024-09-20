package main

import (
	"net/http"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/server"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	storage := collector.NewMemStorage()
	server.Storage = &storage

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		// r.Get("/", server.ListMetrics)
		// r.Route("/", func(r chi.Router) {
		// 	r.Get("/value/{metricType}/{metricName}", server.GetMetric)
		// 	r.Post("/update/{metricType}/{metricName}/{metricValue}", server.UpdateMetric)
		// })
	})

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
