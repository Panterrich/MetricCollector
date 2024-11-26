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
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/server"
)

var (
	DefaultEndPoint = "localhost:8080"
	DefaultLogLevel = zerolog.InfoLevel
)

type Config struct {
	EndPoint string `env:"ADDRESS"`
	LogLevel int    `env:"LOG_LVL"`
}

var (
	cfg    Config
	cfgEnv = Config{
		LogLevel: int(zerolog.TraceLevel) - 1,
	}

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
	root.Flags().IntVar(&cfg.LogLevel, "log-level", int(DefaultLogLevel), "log level (zerolog)")
}

func preRun(_ *cobra.Command, _ []string) {
	if cfgEnv.EndPoint != "" {
		cfg.EndPoint = cfgEnv.EndPoint
	}

	if cfgEnv.LogLevel >= int(zerolog.TraceLevel) {
		cfg.LogLevel = cfgEnv.LogLevel
	}
}

func run(_ *cobra.Command, _ []string) error {
	zerolog.SetGlobalLevel(zerolog.Level(cfg.LogLevel))

	storage := collector.NewMemStorage()
	server.Storage = &storage

	r := chi.NewRouter()

	r.Use(server.WithLogging)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", server.GetListMetrics)
		r.Route("/", func(r chi.Router) {
			r.Post("/value/", server.GetMetricJSON)
			r.Post("/update/", server.UpdateMetricJSON)
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
	err := env.Parse(&cfgEnv)
	if err != nil {
		log.Err(err).Send()
		os.Exit(1)
	}

	err = root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
