package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/server"
	"github.com/Panterrich/MetricCollector/internal/storages"
)

func newCollector(ctx context.Context, cfg Config) (collector.Collector, error) {
	var (
		c   collector.Collector
		err error
		db  *sql.DB
	)

	switch {
	case cfg.DatabaseDsn != "":
		db, err = sql.Open("pgx", cfg.DatabaseDsn)
		if err != nil {
			return nil, fmt.Errorf("database create \"%s\": %w", cfg.DatabaseDsn, err)
		}

		c, err = storages.NewDatabase(ctx, storages.DatabaseParams{
			DB: db,
		})
	case cfg.FileStoragePath != "":
		c, err = storages.NewFile(ctx, storages.FileParams{
			FilePath:      cfg.FileStoragePath,
			Restore:       cfg.Restore,
			StoreInterval: cfg.StoreInterval,
		})
	default:
		c = storages.NewMemory()
	}

	if err != nil {
		c.Close()
		return nil, fmt.Errorf("new collector: %w", err)
	}

	return c, nil
}

func run(_ *cobra.Command, _ []string) error {
	zerolog.SetGlobalLevel(zerolog.Level(cfg.LogLevel))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := newCollector(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create collector with cfg: %w", err)
	}
	defer c.Close()

	r := chi.NewRouter()

	if cfg.KeyHash != "" {
		r.Use(server.WithHashing([]byte(cfg.KeyHash)))
	}

	r.Use(server.WithGzipCompression)
	r.Use(server.WithLogging)

	r.Route("/", func(r chi.Router) {
		r.Get("/", server.WithCollector(c, server.GetListMetrics))
		r.Route("/", func(r chi.Router) {
			r.Get("/ping", server.WithDatabase(c, server.PingDatabase))
			r.Route("/value", func(r chi.Router) {
				r.Post("/", server.WithCollector(c, server.GetMetricJSON))
				r.Get("/{metricType}/{metricName}", server.WithCollector(c, server.GetMetric))
			})
			r.Route("/update", func(r chi.Router) {
				r.Post("/", server.WithCollector(c, server.UpdateMetricJSON))
				r.Post("/{metricType}/{metricName}/{metricValue}", server.WithCollector(c, server.UpdateMetric))
			})
			r.Route("/updates", func(r chi.Router) {
				r.Post("/", server.WithCollector(c, server.UpdateMetricsJSON))
			})
		})
	})

	err = http.ListenAndServe(cfg.EndPoint, r)
	if err != nil {
		return fmt.Errorf("http server internal error: %w", err)
	}

	return nil
}
