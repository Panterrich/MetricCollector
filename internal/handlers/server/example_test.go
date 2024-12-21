package server_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/Panterrich/MetricCollector/internal/handlers/server"
	"github.com/Panterrich/MetricCollector/internal/storages"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

func ExampleGetListMetrics() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := storages.NewMemory()

	ms := []metrics.Metric{
		metrics.NewCounter("counter_1"),
		metrics.NewCounter("counter_2"),
		metrics.NewGauge("gauge_1"),
	}

	ms[0].Update(int64(5))
	ms[1].Update(int64(-1))
	ms[2].Update(float64(1.0))

	c.UpdateMetrics(ctx, ms)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.GetListMetrics(c, w, req)

	resp := w.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))

	// Ordered output:
	// 200
	// text/html
	//
	// Unordered output:
	//  counter_1 (counter): 5
	//  counter_2 (counter): -1
	//    gauge_1 (gauge): 1
}
