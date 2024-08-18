package client

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/metrics"
)

type MemRuntimeStat struct {
	Name   string
	Getter func(m *runtime.MemStats) any
}

var MemRuntimeStats []MemRuntimeStat = []MemRuntimeStat{
	{
		Name:   "Alloc",
		Getter: func(m *runtime.MemStats) any { return m.Alloc },
	},
	{
		Name:   "BuckHashSys",
		Getter: func(m *runtime.MemStats) any { return m.BuckHashSys },
	},
	{
		Name:   "Frees",
		Getter: func(m *runtime.MemStats) any { return m.Frees },
	},
	{
		Name:   "GCCPUFraction",
		Getter: func(m *runtime.MemStats) any { return m.GCCPUFraction },
	},
	{
		Name:   "GCSys",
		Getter: func(m *runtime.MemStats) any { return m.GCSys },
	},
	{
		Name:   "HeapAlloc",
		Getter: func(m *runtime.MemStats) any { return m.HeapAlloc },
	},
	{
		Name:   "HeapIdle",
		Getter: func(m *runtime.MemStats) any { return m.HeapIdle },
	},
	{
		Name:   "HeapInuse",
		Getter: func(m *runtime.MemStats) any { return m.HeapInuse },
	},
	{
		Name:   "HeapObjects",
		Getter: func(m *runtime.MemStats) any { return m.HeapObjects },
	},
	{
		Name:   "HeapReleased",
		Getter: func(m *runtime.MemStats) any { return m.HeapReleased },
	},
	{
		Name:   "HeapSys",
		Getter: func(m *runtime.MemStats) any { return m.HeapSys },
	},
	{
		Name:   "LastGC",
		Getter: func(m *runtime.MemStats) any { return m.LastGC },
	},
	{
		Name:   "Lookups",
		Getter: func(m *runtime.MemStats) any { return m.Lookups },
	},
	{
		Name:   "MCacheInuse",
		Getter: func(m *runtime.MemStats) any { return m.MCacheInuse },
	},
	{
		Name:   "MCacheSys",
		Getter: func(m *runtime.MemStats) any { return m.MCacheSys },
	},
	{
		Name:   "MSpanInuse",
		Getter: func(m *runtime.MemStats) any { return m.MSpanInuse },
	},
	{
		Name:   "MSpanSys",
		Getter: func(m *runtime.MemStats) any { return m.MSpanSys },
	},
	{
		Name:   "Mallocs",
		Getter: func(m *runtime.MemStats) any { return m.Mallocs },
	},
	{
		Name:   "NextGC",
		Getter: func(m *runtime.MemStats) any { return m.NextGC },
	},
	{
		Name:   "NumForcedGC",
		Getter: func(m *runtime.MemStats) any { return m.NumForcedGC },
	},
	{
		Name:   "NumGC",
		Getter: func(m *runtime.MemStats) any { return m.NumGC },
	},
	{
		Name:   "OtherSys",
		Getter: func(m *runtime.MemStats) any { return m.OtherSys },
	},
	{
		Name:   "PauseTotalNs",
		Getter: func(m *runtime.MemStats) any { return m.PauseTotalNs },
	},
	{
		Name:   "StackInuse",
		Getter: func(m *runtime.MemStats) any { return m.StackInuse },
	},
	{
		Name:   "StackSys",
		Getter: func(m *runtime.MemStats) any { return m.StackSys },
	},
	{
		Name:   "Sys",
		Getter: func(m *runtime.MemStats) any { return m.Sys },
	},
	{
		Name:   "TotalAlloc",
		Getter: func(m *runtime.MemStats) any { return m.TotalAlloc },
	},
}

func UpdateAllMetrics(storage collector.Collector) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	for _, v := range MemRuntimeStats {
		storage.UpdateMetric(metrics.TypeMetricGauge, v.Name, v.Getter(&memStats))
	}

	storage.UpdateMetric(metrics.TypeMetricCounter, "PollCount", 1)
	storage.UpdateMetric(metrics.TypeMetricGauge, "RandomValue", rand.Float64())
}

func ReportAllMetrics(storage collector.Collector, client *http.Client) {
	metrics := storage.GetAllMetrics()
	for _, metric := range metrics {
		ReportMetric(metric, client)
	}
}

func ReportMetric(metric metrics.Metric, client *http.Client) {
	url := "/update/" + metric.Type() + "/" + metric.Name() + "/"

	switch metric.Type() {
	case metrics.TypeMetricCounter:
		val, ok := metric.Value().(int64)
		if !ok {
			return
		}
		url += strconv.FormatInt(val, 10)

	case metrics.TypeMetricGauge:
		val, ok := metric.Value().(float64)
		if !ok {
			return
		}
		url += strconv.FormatFloat(val, 'f', -1, 64)

	default:
		fmt.Println("unknown type")
		return
	}

	request, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		return
	}

	response, err := client.Do(request)
	if err != nil {
		return
	}

	defer response.Body.Close()

	_, err = io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
}
