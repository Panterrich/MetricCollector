package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/agent"
)

var (
	UpdateInterval time.Duration = 2 * time.Second
	ReportInterval time.Duration = 10 * time.Second
)

func main() {
	var metrics collector.Collector
	storage := collector.NewMemStorage()
	metrics = &storage

	client := resty.New()
	serverAddress := "localhost:8080"

	updateTimer := time.NewTicker(UpdateInterval)
	reportTimer := time.NewTicker(ReportInterval)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-stop:
			return
		case <-updateTimer.C:
			agent.UpdateAllMetrics(metrics)
		case <-reportTimer.C:
			agent.ReportAllMetrics(metrics, client, serverAddress)
		}
	}
}
