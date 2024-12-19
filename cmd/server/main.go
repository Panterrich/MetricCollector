package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/server"
	"github.com/Panterrich/MetricCollector/internal/storages"
)

var (
	DefaultEndPoint             = "localhost:8080"
	DefaultLogLevel             = zerolog.InfoLevel
	DefaultStoreInterval   uint = 300
	DefaultFileStoragePath      = ""
	DefaultRestore              = true
	DefaultDatabaseDsn          = ""
	DefaultKeyHash              = ""
)

type Config struct {
	EndPoint        string `env:"ADDRESS"`
	LogLevel        int    `env:"LOG_LVL"`
	StoreInterval   uint   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	KeyHash         string `env:"KEY"`
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
	root.Flags().UintVarP(&cfg.StoreInterval, "i", "i", DefaultStoreInterval, "store interval")
	root.Flags().StringVarP(&cfg.FileStoragePath, "f", "f", DefaultFileStoragePath, "file storage path")
	root.Flags().BoolVarP(&cfg.Restore, "r", "r", DefaultRestore, "restore")
	root.Flags().StringVarP(&cfg.DatabaseDsn, "d", "d", DefaultDatabaseDsn, "database dsn")
	root.Flags().StringVarP(&cfg.KeyHash, "key", "k", DefaultKeyHash, "key for hash sha256")
}

func preRun(_ *cobra.Command, _ []string) {
	if cfgEnv.EndPoint != "" {
		cfg.EndPoint = cfgEnv.EndPoint
	}

	if cfgEnv.LogLevel >= int(zerolog.TraceLevel) {
		cfg.LogLevel = cfgEnv.LogLevel
	}

	if cfgEnv.StoreInterval != 0 {
		cfg.StoreInterval = cfgEnv.StoreInterval
	}

	if cfgEnv.FileStoragePath != "" {
		cfg.FileStoragePath = cfgEnv.FileStoragePath
	}

	if cfgEnv.Restore {
		cfg.Restore = cfgEnv.Restore
	}

	if cfgEnv.DatabaseDsn != "" {
		cfg.DatabaseDsn = cfgEnv.DatabaseDsn
	}

	if cfgEnv.KeyHash != "" {
		cfg.KeyHash = cfgEnv.KeyHash
	}
}

func run(_ *cobra.Command, _ []string) error {
	zerolog.SetGlobalLevel(zerolog.Level(cfg.LogLevel))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := NewCollector(ctx, cfg)
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

func NewCollector(ctx context.Context, cfg Config) (collector.Collector, error) {
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
