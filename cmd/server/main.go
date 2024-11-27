package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/server"
	"github.com/Panterrich/MetricCollector/pkg/serialization"
)

var (
	DefaultEndPoint             = "localhost:8080"
	DefaultLogLevel             = zerolog.InfoLevel
	DefaultStoreInterval   uint = 300
	DefaultFileStoragePath      = ""
	DefaultRestore              = true
)

type Config struct {
	EndPoint        string `env:"ADDRESS"`
	LogLevel        int    `env:"LOG_LVL"`
	StoreInterval   uint   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
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
}

func run(_ *cobra.Command, _ []string) error {
	zerolog.SetGlobalLevel(zerolog.Level(cfg.LogLevel))

	storage := collector.NewMemStorage()
	server.Storage = &storage

	if cfg.Restore && cfg.FileStoragePath != "" {
		if err := serialization.Load(&storage, cfg.FileStoragePath); err != nil {
			return fmt.Errorf("can't load database: %w", err)
		}
	}

	r := chi.NewRouter()

	r.Use(server.WithLogging)
	r.Use(server.WithGzipCompression)

	r.Route("/", func(r chi.Router) {
		r.Get("/", server.GetListMetrics)
		r.Route("/", func(r chi.Router) {
			r.Route("/value", func(r chi.Router) {
				r.Post("/", server.GetMetricJSON)
				r.Get("/{metricType}/{metricName}", server.GetMetric)
			})
			r.Route("/update", func(r chi.Router) {
				r.Post("/", server.UpdateMetricJSON)
				r.Post("/{metricType}/{metricName}/{metricValue}", server.UpdateMetric)

				if cfg.StoreInterval == 0 {
					r.Use(server.WithBackup(&storage, cfg.FileStoragePath))
				}
			})
		})
	})

	var backupTimer *time.Ticker

	if cfg.StoreInterval != 0 {
		backupTimer = time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
	}

	stop := make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-backupTimer.C:
				if err := serialization.Save(&storage, cfg.FileStoragePath); err != nil {
					log.Error().Msgf("can't save database: %v", err)
				}
			case <-stop:
				return
			}
		}
	}()

	err := http.ListenAndServe(cfg.EndPoint, r)
	if err != nil {
		return fmt.Errorf("http server internal error: %w", err)
	}

	stop <- struct{}{}

	backupTimer.Stop()
	wg.Wait()

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
