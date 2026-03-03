package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Panterrich/MetricCollector/internal/handlers/agent"
	"github.com/Panterrich/MetricCollector/internal/storages"
	runtime_stats "github.com/Panterrich/MetricCollector/pkg/runtime-stats"
	"github.com/Panterrich/MetricCollector/pkg/workpool"
)

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
