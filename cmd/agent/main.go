package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/agent"

	"github.com/spf13/cobra"
)

var (
	DefaultEndPoint       string = "localhost:8080"
	DefaultReportInterval uint   = 10
	DefaultPollInterval   uint   = 2
)

var (
	flagEndPoint       string
	flagReportInterval uint
	flagPollInterval   uint

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

			if flagPollInterval == 0 || flagReportInterval == 0 {
				return fmt.Errorf("zero interval")
			}

			return nil
		},
		RunE: run,
	}
)

func init() {
	root.Flags().StringVarP(&flagEndPoint, "a", "a", "localhost:8080", "end-point for HTTP-server")
	root.Flags().UintVarP(&flagReportInterval, "r", "r", 10, "report interval")
	root.Flags().UintVarP(&flagPollInterval, "p", "p", 2, "poll interval")
}

func run(cmd *cobra.Command, args []string) error {
	var metrics collector.Collector
	storage := collector.NewMemStorage()
	metrics = &storage

	client := resty.New()
	serverAddress := flagEndPoint

	reportTimer := time.NewTicker(time.Duration(flagReportInterval))
	pollTimer := time.NewTicker(time.Duration(flagPollInterval))

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
	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
