package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/server"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"github.com/spf13/cobra"
)

var (
	DefaultEndPoint string = "localhost:8080"
)

var (
	flagEndPoint string

	root = &cobra.Command{
		Use:   "server",
		Short: "Server for storing metrics",
		Long:  "Server for storing metrics",
		Args: func(cmd *cobra.Command, args []string) error {

			if err := cobra.ExactArgs(0)(cmd, args); err != nil {
				return err
			}

			if _, _, err := net.SplitHostPort(flagEndPoint); err != nil {
				return fmt.Errorf("invalid end-point for HTTP-server: %w", err)
			}

			return nil
		},
		RunE: run,
	}
)

func init() {
	root.Flags().StringVarP(&flagEndPoint, "a", "a", DefaultEndPoint, "end-point for HTTP-server")
}

func run(cmd *cobra.Command, args []string) error {
	storage := collector.NewMemStorage()
	server.Storage = &storage

	r := chi.NewRouter()

	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", server.GetListMetrics)
		r.Route("/", func(r chi.Router) {
			r.Get("/value/{metricType}/{metricName}", server.GetMetric)
			r.Post("/update/{metricType}/{metricName}/{metricValue}", server.UpdateMetric)
		})
	})

	err := http.ListenAndServe(flagEndPoint, r)
	if err != nil {
		return fmt.Errorf("http server internal error: %w", err)
	}

	return nil
}

func main() {
	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
