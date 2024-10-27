package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/agent"
)

var (
	DefaultEndPoint            = "localhost:8080"
	DefaultReportInterval uint = 10
	DefaultPollInterval   uint = 2
)

type Config struct {
	EndPoint       string `env:"ADDRESS"`
	ReportInterval uint   `env:"REPORT_INTERVAL"`
	PollInterval   uint   `env:"POLL_INTERVAL"`
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

			if cfg.PollInterval == 0 || cfg.ReportInterval == 0 {
				return fmt.Errorf("zero interval")
			}

			return nil
		},
		PreRun: preRun,
		RunE:   run,
	}
)

func init() {
	root.Flags().StringVarP(&cfg.EndPoint, "a", "a", "localhost:8080", "end-point for HTTP-server")
	root.Flags().UintVarP(&cfg.ReportInterval, "r", "r", 10, "report interval")
	root.Flags().UintVarP(&cfg.PollInterval, "p", "p", 2, "poll interval")
}

func preRun(_ *cobra.Command, _ []string) {
	if cfgEnv.EndPoint != "" {
		cfg.EndPoint = cfgEnv.EndPoint
	}

	if cfgEnv.ReportInterval != 0 {
		cfg.ReportInterval = cfgEnv.ReportInterval
	}

	if cfgEnv.PollInterval != 0 {
		cfg.PollInterval = cfgEnv.PollInterval
	}
}

func run(_ *cobra.Command, _ []string) error {
	var metrics collector.Collector

	storage := collector.NewMemStorage()
	metrics = &storage

	client := resty.New()
	serverAddress := cfg.EndPoint

	reportTimer := time.NewTicker(time.Duration(cfg.ReportInterval))
	pollTimer := time.NewTicker(time.Duration(cfg.PollInterval))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-stop:
			return nil
		case <-reportTimer.C:
			agent.ReportAllMetrics(metrics, client, serverAddress)
		case <-pollTimer.C:
			agent.UpdateAllMetrics(metrics)
		}
	}
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
