package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/server"
)

var (
	DefaultEndPoint = "localhost:8080"
)

type Config struct {
	EndPoint string `env:"ADDRESS"`
}

var (
	cfgEnv Config
	cfg    Config

	root = &cobra.Command{
		Use:   "server",
		Short: "Server for storing metrics",
		Long:  "Server for storing metrics",
		Args: func(cmd *cobra.Command, args []string) error {

			if err := cobra.ExactArgs(0)(cmd, args); err != nil {
				return err
			}

			if _, _, err := net.SplitHostPort(cfg.EndPoint); err != nil {
				return fmt.Errorf("invalid end-point for HTTP-server: %w", err)
			}

			return nil
		},
		PreRun: preRun,
		RunE:   run,
	}
)

func init() {
	root.Flags().StringVarP(&cfg.EndPoint, "a", "a", DefaultEndPoint, "end-point for HTTP-server")
}

func preRun(_ *cobra.Command, _ []string) {
	if cfgEnv.EndPoint != "" {
		cfg.EndPoint = cfgEnv.EndPoint
	}
}

func run(_ *cobra.Command, _ []string) error {
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

	err := http.ListenAndServe(cfg.EndPoint, r)
	if err != nil {
		return fmt.Errorf("http server internal error: %w", err)
	}

	return nil
}

func main() {
	logger := zerolog.Logger{}

	err := env.Parse(&cfgEnv)
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}

	err = root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
