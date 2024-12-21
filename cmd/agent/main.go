package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Panterrich/MetricCollector/internal/handlers/agent"
	"github.com/Panterrich/MetricCollector/internal/storages"
	runtime_stats "github.com/Panterrich/MetricCollector/pkg/runtime-stats"
	"github.com/Panterrich/MetricCollector/pkg/workpool"
)

var (
	DefaultEndPoint            = "localhost:8080"
	DefaultReportInterval uint = 10
	DefaultPollInterval   uint = 2
	DefaultKeyHash             = ""
	DefaultRateLimit      uint = 1
)

type Config struct {
	EndPoint       string `env:"ADDRESS"`
	ReportInterval uint   `env:"REPORT_INTERVAL"`
	PollInterval   uint   `env:"POLL_INTERVAL"`
	KeyHash        string `env:"KEY"`
	RateLimit      uint   `env:"RATE_LIMIT"`
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

			if cfg.RateLimit == 0 {
				return fmt.Errorf("zero rate limit")
			}

			return nil
		},
		PreRun: preRun,
		RunE:   run,
	}
)

func init() {
	root.Flags().StringVarP(&cfg.EndPoint, "a", "a", DefaultEndPoint, "end-point for HTTP-server")
	root.Flags().UintVarP(&cfg.ReportInterval, "r", "r", DefaultReportInterval, "report interval")
	root.Flags().UintVarP(&cfg.PollInterval, "p", "p", DefaultPollInterval, "poll interval")
	root.Flags().StringVarP(&cfg.KeyHash, "key", "k", DefaultKeyHash, "key for hash sha256")
	root.Flags().UintVarP(&cfg.RateLimit, "l", "l", DefaultRateLimit, "rate limit")
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

	if cfgEnv.KeyHash != "" {
		cfg.KeyHash = cfgEnv.KeyHash
	}
}

func run(_ *cobra.Command, _ []string) error {
	storage := storages.NewMemory()

	client := resty.New()
	serverAddress := cfg.EndPoint

	reportTimer := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	pollTimer := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := workpool.NewPool(ctx, int(cfg.RateLimit))

	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		defer wg.Done()

		<-stop
		cancel()
	}()

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case res := <-pool.Results:
				log.Debug().Err(res.Err).Msg(res.Msg)
			}
		}
	}()

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-reportTimer.C:
				pool.Schedule(ctx, func(ctx context.Context) workpool.Result {
					agent.ReportAllMetrics(ctx, storage, client, serverAddress, cfg.KeyHash)

					return workpool.Result{
						Msg: "report all",
						Err: nil,
					}
				})
			case <-pollTimer.C:
				runtime_stats.UpdateAllMetrics(ctx, pool, storage)
			}
		}
	}()

	pool.Wait()
	wg.Wait()

	return nil
}

func main() {
	err := env.Parse(&cfgEnv)
	if err != nil {
		log.Err(err).Send()
		return
	}

	err = root.Execute()
	if err != nil {
		log.Err(err).Send()
		return
	}
}
