package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.openly.dev/pointy"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/handlers/server"
	"github.com/Panterrich/MetricCollector/internal/storages"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
	"github.com/Panterrich/MetricCollector/pkg/serialization"
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

	t.Run("ExistedCounter", func(t *testing.T) {
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

	t.Run("ExistedGauge", func(t *testing.T) {
		w := httptest.NewRecorder()
		req :=
			WithURLParams(
				httptest.NewRequest(http.MethodGet, "/value/gauge/gauge_1", nil),
				map[string]string{
					"metricType": "gauge",
					"metricName": "gauge_1",
				})

		server.WithCollector(c, server.GetMetric)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		assert.Equal(t, string(body), fmt.Sprintf("%v\n", 1.0))
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

func TestUpdateMetric(t *testing.T) {
	c := GetBaseCollector()

	t.Run("ExistedCounter", func(t *testing.T) {
		w := httptest.NewRecorder()
		req :=
			WithURLParams(
				httptest.NewRequest(http.MethodPost, "/update/counter/counter_1/10", nil),
				map[string]string{
					"metricType":  "counter",
					"metricName":  "counter_1",
					"metricValue": "10",
				})

		server.WithCollector(c, server.UpdateMetric)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		val, err := c.GetMetric(context.TODO(), metrics.TypeMetricCounter, "counter_1")
		assert.NoError(t, err)
		assert.Equal(t, int64(15), val.(int64))
	})

	t.Run("NewCounter", func(t *testing.T) {
		w := httptest.NewRecorder()
		req :=
			WithURLParams(
				httptest.NewRequest(http.MethodPost, "/update/counter/counter_3/10", nil),
				map[string]string{
					"metricType":  "counter",
					"metricName":  "counter_3",
					"metricValue": "10",
				})

		server.WithCollector(c, server.UpdateMetric)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		val, err := c.GetMetric(context.TODO(), metrics.TypeMetricCounter, "counter_3")
		assert.NoError(t, err)
		assert.Equal(t, int64(10), val.(int64))
	})

	t.Run("ExistedGauge", func(t *testing.T) {
		w := httptest.NewRecorder()
		req :=
			WithURLParams(
				httptest.NewRequest(http.MethodPost, "/update/gauge/gauge_1/1.0", nil),
				map[string]string{
					"metricType":  "gauge",
					"metricName":  "gauge_1",
					"metricValue": "1.0",
				})

		server.WithCollector(c, server.UpdateMetric)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		val, err := c.GetMetric(context.TODO(), metrics.TypeMetricGauge, "gauge_1")
		assert.NoError(t, err)
		assert.Equal(t, float64(1.0), val.(float64))
	})

	t.Run("NewGauge", func(t *testing.T) {
		w := httptest.NewRecorder()
		req :=
			WithURLParams(
				httptest.NewRequest(http.MethodPost, "/update/gauge/gauge_3/5.0", nil),
				map[string]string{
					"metricType":  "gauge",
					"metricName":  "gauge_3",
					"metricValue": "5.0",
				})

		server.WithCollector(c, server.UpdateMetric)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		val, err := c.GetMetric(context.TODO(), metrics.TypeMetricGauge, "gauge_3")
		assert.NoError(t, err)
		assert.Equal(t, float64(5.0), val.(float64))
	})

	t.Run("InvalidType", func(t *testing.T) {
		w := httptest.NewRecorder()
		req :=
			WithURLParams(
				httptest.NewRequest(http.MethodPost, "/update/gauger/gauge_1/1.0", nil),
				map[string]string{
					"metricType":  "gauger",
					"metricName":  "gauge_1",
					"metricValue": "1.0",
				})

		server.WithCollector(c, server.UpdateMetric)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("InvalidValue", func(t *testing.T) {
		w := httptest.NewRecorder()
		req :=
			WithURLParams(
				httptest.NewRequest(http.MethodPost, "/update/gauger/counter_2/1.0", nil),
				map[string]string{
					"metricType":  "counter",
					"metricName":  "counter_2",
					"metricValue": "1.0",
				})

		server.WithCollector(c, server.UpdateMetric)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestGetMetricJSON(t *testing.T) {
	c := GetBaseCollector()

	t.Run("ExistedCounter", func(t *testing.T) {
		metric := serialization.Metric{
			ID:    "counter_1",
			MType: "counter",
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/value/", &body)

		server.WithCollector(c, server.GetMetricJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&metric))

		assert.Equal(t, serialization.Metric{
			ID:    "counter_1",
			MType: "counter",
			Delta: pointy.Int64(5),
		}, metric)
	})

	t.Run("ExistedGauge", func(t *testing.T) {
		metric := serialization.Metric{
			ID:    "gauge_1",
			MType: "gauge",
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/value/", &body)

		server.WithCollector(c, server.GetMetricJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&metric))

		assert.Equal(t, serialization.Metric{
			ID:    "gauge_1",
			MType: "gauge",
			Val:   pointy.Float64(1.0),
		}, metric)
	})

	t.Run("NotFound", func(t *testing.T) {
		metric := serialization.Metric{
			ID:    "gauge_3",
			MType: "gauge",
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/value/", &body)

		server.WithCollector(c, server.GetMetricJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("InvalidJson", func(t *testing.T) {

		var body bytes.Buffer

		body.Write([]byte("string"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/value/", &body)

		server.WithCollector(c, server.GetMetricJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("InvalidType", func(t *testing.T) {
		metric := serialization.Metric{
			ID:    "gauge_1",
			MType: "gauger",
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/value/", &body)

		server.WithCollector(c, server.GetMetricJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("InvalidWrite", func(t *testing.T) {
		metric := serialization.Metric{
			ID:    "gauge_1",
			MType: "gauge",
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := &MockHTTPResponseWriter{}
		req := httptest.NewRequest(http.MethodPost, "/value/", &body)

		server.WithCollector(c, server.GetMetricJSON)(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.status)
	})
}

func TestUpdateMetricJSON(t *testing.T) {
	c := GetBaseCollector()

	t.Run("ExistedCounter", func(t *testing.T) {
		metric := serialization.Metric{
			ID:    "counter_1",
			MType: "counter",
			Delta: pointy.Int64(10),
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/update/", &body)

		server.WithCollector(c, server.UpdateMetricJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&metric))

		assert.Equal(t, serialization.Metric{
			ID:    "counter_1",
			MType: "counter",
			Delta: pointy.Int64(15),
		}, metric)
	})

	t.Run("NewCounter", func(t *testing.T) {
		metric := serialization.Metric{
			ID:    "counter_3",
			MType: "counter",
			Delta: pointy.Int64(10),
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/update/", &body)

		server.WithCollector(c, server.UpdateMetricJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&metric))

		assert.Equal(t, serialization.Metric{
			ID:    "counter_3",
			MType: "counter",
			Delta: pointy.Int64(10),
		}, metric)
	})

	t.Run("ExistedGauge", func(t *testing.T) {
		metric := serialization.Metric{
			ID:    "gauge_1",
			MType: "gauge",
			Val:   pointy.Float64(5.0),
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/update/", &body)

		server.WithCollector(c, server.UpdateMetricJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&metric))

		assert.Equal(t, serialization.Metric{
			ID:    "gauge_1",
			MType: "gauge",
			Val:   pointy.Float64(5.0),
		}, metric)
	})

	t.Run("NewGauge", func(t *testing.T) {
		metric := serialization.Metric{
			ID:    "gauge_3",
			MType: "gauge",
			Val:   pointy.Float64(7.0),
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/update/", &body)

		server.WithCollector(c, server.UpdateMetricJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&metric))

		assert.Equal(t, serialization.Metric{
			ID:    "gauge_3",
			MType: "gauge",
			Val:   pointy.Float64(7.0),
		}, metric)
	})

	t.Run("InvalidJson", func(t *testing.T) {
		var body bytes.Buffer

		body.Write([]byte("string"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/update/", &body)

		server.WithCollector(c, server.UpdateMetricJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("InvalidType", func(t *testing.T) {
		metric := serialization.Metric{
			ID:    "gauge_3",
			MType: "gauger",
			Val:   pointy.Float64(7.0),
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/update/", &body)

		server.WithCollector(c, server.UpdateMetricJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("InvalidValue", func(t *testing.T) {
		metric := serialization.Metric{
			ID:    "gauge_1",
			MType: "gauge",
			Delta: pointy.Int64(10),
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/update/", &body)

		server.WithCollector(c, server.UpdateMetricJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("InvalidWrite", func(t *testing.T) {
		metric := serialization.Metric{
			ID:    "gauge_3",
			MType: "gauge",
			Val:   pointy.Float64(7.0),
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := &MockHTTPResponseWriter{}
		req := httptest.NewRequest(http.MethodPost, "/update/", &body)

		server.WithCollector(c, server.UpdateMetricJSON)(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.status)
	})
}

func TestUpdateMetricsJSON(t *testing.T) {
	c := GetBaseCollector()

	t.Run("SeveralMetrics", func(t *testing.T) {
		metric := []serialization.Metric{
			{
				ID:    "counter_1",
				MType: "counter",
				Delta: pointy.Int64(10),
			},
			{
				ID:    "gauge_1",
				MType: "gauge",
				Val:   pointy.Float64(1.0),
			},
			{
				ID:    "counter_3",
				MType: "counter",
				Delta: pointy.Int64(4),
			},
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/updates/", &body)

		server.WithCollector(c, server.UpdateMetricsJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(req.Body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		expected, err := serialization.ConvertToJSONMetrics(c.GetAllMetrics(context.TODO()))
		assert.NoError(t, err)

		assert.ElementsMatch(t, []serialization.Metric{
			{
				ID:    "counter_1",
				MType: "counter",
				Delta: pointy.Int64(15),
			},
			{
				ID:    "counter_2",
				MType: "counter",
				Delta: pointy.Int64(-1),
			},
			{
				ID:    "gauge_1",
				MType: "gauge",
				Val:   pointy.Float64(1.0),
			},
			{
				ID:    "counter_3",
				MType: "counter",
				Delta: pointy.Int64(4),
			},
		}, expected)
	})

	t.Run("InvalidJson", func(t *testing.T) {
		var body bytes.Buffer

		body.Write([]byte("string"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/updates/", &body)

		server.WithCollector(c, server.UpdateMetricsJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("InvalidType", func(t *testing.T) {
		metric := []serialization.Metric{
			{
				ID:    "gauge_3",
				MType: "gauger",
				Val:   pointy.Float64(7.0),
			},
		}

		var body bytes.Buffer

		assert.NoError(t, json.NewEncoder(&body).Encode(&metric))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/updates/", &body)

		server.WithCollector(c, server.UpdateMetricsJSON)(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		_, _ = io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
