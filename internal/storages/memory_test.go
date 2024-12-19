package storages_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/Panterrich/MetricCollector/internal/storages"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

const (
	Count = 3

	NConsumers = 10
	NProducers = 100
	Attempts   = 10000
)

func TestMemory_UpdateSequenceCounter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := storages.NewMemory()

	_, err := m.GetMetric(ctx, metrics.TypeMetricCounter, "counter")
	assert.Error(t, err)

	var a int64

	for i := 0; i < Count; i++ {
		v := rand.Int64N(1024)
		a += v

		assert.NoError(t, m.UpdateMetric(ctx, metrics.TypeMetricCounter, "counter", v))

		val, err := m.GetMetric(ctx, metrics.TypeMetricCounter, "counter")
		assert.NoError(t, err)
		assert.Equal(t, a, val.(int64))
	}
}

func TestMemory_UpdateSequenceGauge(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := storages.NewMemory()

	_, err := m.GetMetric(ctx, metrics.TypeMetricGauge, "gauge")
	assert.Error(t, err)

	for i := 0; i < Count; i++ {
		v := rand.Float64()

		assert.NoError(t, m.UpdateMetric(ctx, metrics.TypeMetricGauge, "gauge", v))

		val, err := m.GetMetric(ctx, metrics.TypeMetricGauge, "gauge")
		assert.NoError(t, err)
		assert.Equal(t, v, val.(float64))
	}
}

func TestMemory_SingleCounter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := storages.NewMemory()

	var wg sync.WaitGroup

	wg.Add(NConsumers + NProducers)

	for i := 0; i < NProducers; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < Attempts; j++ {
				m.UpdateMetric(ctx, metrics.TypeMetricCounter, "counter", int64(1))
			}
		}()
	}

	for i := 0; i < NConsumers; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < Attempts; j++ {
				m.GetAllMetrics(ctx)
			}
		}()
	}

	wg.Wait()

	val, err := m.GetMetric(ctx, metrics.TypeMetricCounter, "counter")

	assert.Nil(t, err)
	assert.Equal(t, int64(NProducers*Attempts), val.(int64))

	m.Close()
}

func TestMemory_SeveralTypes(_ *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := storages.NewMemory()

	var wg sync.WaitGroup

	wg.Add(NConsumers + NProducers)

	for i := 0; i < NProducers; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < Attempts; j++ {
				if j%2 == 0 {
					m.UpdateMetric(ctx, metrics.TypeMetricCounter, fmt.Sprintf("counter_%d", j%5), int64(1))
				} else {
					m.UpdateMetric(ctx, metrics.TypeMetricGauge, fmt.Sprintf("gauge_%d", j%5), rand.Float64())
				}
			}
		}()
	}

	for i := 0; i < NConsumers; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < Attempts; j++ {
				m.GetAllMetrics(ctx)
			}
		}()
	}

	wg.Wait()
	m.Close()
}

func BenchmarkMemory(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < b.N; i++ {
		m := storages.NewMemory()

		var wg sync.WaitGroup

		wg.Add(NConsumers + NProducers)

		for i := 0; i < NProducers; i++ {
			go func() {
				defer wg.Done()

				for j := 0; j < Attempts; j++ {
					m.UpdateMetric(ctx, metrics.TypeMetricCounter, "counter", int64(1))
				}
			}()
		}

		for i := 0; i < NConsumers; i++ {
			go func() {
				defer wg.Done()

				for j := 0; j < Attempts; j++ {
					m.GetAllMetrics(ctx)
				}
			}()
		}

		wg.Wait()
	}
}
