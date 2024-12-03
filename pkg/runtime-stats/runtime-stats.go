package runtimestats

import (
	"context"
	"math/rand/v2"
	"runtime"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

type MemRuntimeStat struct {
	Name   string
	Getter func(m *runtime.MemStats) any
}

var MemRuntimeStats = []MemRuntimeStat{
	{
		Name:   "Alloc",
		Getter: func(m *runtime.MemStats) any { return float64(m.Alloc) },
	},
	{
		Name:   "BuckHashSys",
		Getter: func(m *runtime.MemStats) any { return float64(m.BuckHashSys) },
	},
	{
		Name:   "Frees",
		Getter: func(m *runtime.MemStats) any { return float64(m.Frees) },
	},
	{
		Name:   "GCCPUFraction",
		Getter: func(m *runtime.MemStats) any { return float64(m.GCCPUFraction) },
	},
	{
		Name:   "GCSys",
		Getter: func(m *runtime.MemStats) any { return float64(m.GCSys) },
	},
	{
		Name:   "HeapAlloc",
		Getter: func(m *runtime.MemStats) any { return float64(m.HeapAlloc) },
	},
	{
		Name:   "HeapIdle",
		Getter: func(m *runtime.MemStats) any { return float64(m.HeapIdle) },
	},
	{
		Name:   "HeapInuse",
		Getter: func(m *runtime.MemStats) any { return float64(m.HeapInuse) },
	},
	{
		Name:   "HeapObjects",
		Getter: func(m *runtime.MemStats) any { return float64(m.HeapObjects) },
	},
	{
		Name:   "HeapReleased",
		Getter: func(m *runtime.MemStats) any { return float64(m.HeapReleased) },
	},
	{
		Name:   "HeapSys",
		Getter: func(m *runtime.MemStats) any { return float64(m.HeapSys) },
	},
	{
		Name:   "LastGC",
		Getter: func(m *runtime.MemStats) any { return float64(m.LastGC) },
	},
	{
		Name:   "Lookups",
		Getter: func(m *runtime.MemStats) any { return float64(m.Lookups) },
	},
	{
		Name:   "MCacheInuse",
		Getter: func(m *runtime.MemStats) any { return float64(m.MCacheInuse) },
	},
	{
		Name:   "MCacheSys",
		Getter: func(m *runtime.MemStats) any { return float64(m.MCacheSys) },
	},
	{
		Name:   "MSpanInuse",
		Getter: func(m *runtime.MemStats) any { return float64(m.MSpanInuse) },
	},
	{
		Name:   "MSpanSys",
		Getter: func(m *runtime.MemStats) any { return float64(m.MSpanSys) },
	},
	{
		Name:   "Mallocs",
		Getter: func(m *runtime.MemStats) any { return float64(m.Mallocs) },
	},
	{
		Name:   "NextGC",
		Getter: func(m *runtime.MemStats) any { return float64(m.NextGC) },
	},
	{
		Name:   "NumForcedGC",
		Getter: func(m *runtime.MemStats) any { return float64(m.NumForcedGC) },
	},
	{
		Name:   "NumGC",
		Getter: func(m *runtime.MemStats) any { return float64(m.NumGC) },
	},
	{
		Name:   "OtherSys",
		Getter: func(m *runtime.MemStats) any { return float64(m.OtherSys) },
	},
	{
		Name:   "PauseTotalNs",
		Getter: func(m *runtime.MemStats) any { return float64(m.PauseTotalNs) },
	},
	{
		Name:   "StackInuse",
		Getter: func(m *runtime.MemStats) any { return float64(m.StackInuse) },
	},
	{
		Name:   "StackSys",
		Getter: func(m *runtime.MemStats) any { return float64(m.StackSys) },
	},
	{
		Name:   "Sys",
		Getter: func(m *runtime.MemStats) any { return float64(m.Sys) },
	},
	{
		Name:   "TotalAlloc",
		Getter: func(m *runtime.MemStats) any { return float64(m.TotalAlloc) },
	},
}

func UpdateAllMetrics(ctx context.Context, storage collector.Collector) {
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)

	for _, v := range MemRuntimeStats {
		storage.UpdateMetric(ctx, metrics.TypeMetricGauge, v.Name, v.Getter(&memStats))
	}

	storage.UpdateMetric(ctx, metrics.TypeMetricCounter, "PollCount", int64(1))
	storage.UpdateMetric(ctx, metrics.TypeMetricGauge, "RandomValue", rand.Float64())
}
