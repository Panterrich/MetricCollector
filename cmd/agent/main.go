package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/client"
)

var (
	UpdateInterval time.Duration = 2 * time.Second
	ReportInterval time.Duration = 10 * time.Second
)

func main() {
	var metrics collector.Collector
	storage := collector.NewMemStorage()
	metrics = &storage

	httpClient := &http.Client{}

	updateTimer := time.NewTicker(UpdateInterval)
	reportTimer := time.NewTicker(ReportInterval)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-stop:
		return
	case <-updateTimer.C:
		client.UpdateAllMetrics(metrics)
	case <-reportTimer.C:
		client.ReportAllMetrics(metrics, httpClient)
	}
}
