package server_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/server"
	"github.com/Panterrich/MetricCollector/internal/storages"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

func GetBaseCollector() collector.Collector {
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

	return c
}

type MockHTTPResponseWriter struct {
	status int
}

var _ http.ResponseWriter = &MockHTTPResponseWriter{}

func (m *MockHTTPResponseWriter) Header() http.Header {
	return http.Header{}
}

func (m *MockHTTPResponseWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("invalid write")
}

func (m *MockHTTPResponseWriter) WriteHeader(status int) {
	m.status = status
}

func WithURLParams(r *http.Request, params map[string]string) *http.Request {
	chiCtx := chi.NewRouteContext()
	req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))

	for key, value := range params {
		chiCtx.URLParams.Add(key, value)
	}

	return req
}

func TestGetListMetrics(t *testing.T) {
	c := GetBaseCollector()

	t.Run("AllMetrics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		server.WithCollector(c, server.GetListMetrics)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		expected := []string{
			" counter_1 (counter): 5",
			" counter_2 (counter): -1",
			"   gauge_1 (gauge): 1",
		}

		var lines []string

		for _, line := range strings.Split(string(body), "\n") {
			if line != "" {
				lines = append(lines, line)
			}
		}

		assert.ElementsMatch(t, expected, lines)
	})

	t.Run("InvalidWrite", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := &MockHTTPResponseWriter{}

		server.WithCollector(c, server.GetListMetrics)(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.status)
	})
}

func TestGetMetric(t *testing.T) {
	c := GetBaseCollector()

	t.Run("Found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req :=
			WithURLParams(
				httptest.NewRequest(http.MethodGet, "/value/counter/counter_1", nil),
				map[string]string{
					"metricType": "counter",
					"metricName": "counter_1",
				})

		server.WithCollector(c, server.GetMetric)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		assert.Equal(t, string(body), fmt.Sprintf("%v\n", 5))
	})

	t.Run("NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		req :=
			WithURLParams(
				httptest.NewRequest(http.MethodGet, "/value/gauge/counter_1", nil),
				map[string]string{
					"metricType": "gauge",
					"metricName": "counter_1",
				})

		server.WithCollector(c, server.GetMetric)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("InvalidType", func(t *testing.T) {
		w := httptest.NewRecorder()
		req :=
			WithURLParams(
				httptest.NewRequest(http.MethodGet, "/value/gauger/gauge_1", nil),
				map[string]string{
					"metricType": "gauger",
					"metricName": "gauge_1",
				})

		server.WithCollector(c, server.GetMetric)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("InvalidWrite", func(t *testing.T) {
		w := &MockHTTPResponseWriter{}
		req :=
			WithURLParams(
				httptest.NewRequest(http.MethodGet, "/value/gauge/gauge_1", nil),
				map[string]string{
					"metricType": "gauge",
					"metricName": "gauge_1",
				})

		server.WithCollector(c, server.GetMetric)(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.status)
	})
}
