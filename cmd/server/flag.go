package main

import (
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
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
