package storages_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Panterrich/MetricCollector/internal/storages"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

func TestFile_UpdateSequenceCounter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempFile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)

	defer os.Remove(tempFile.Name())

	c, err := storages.NewFile(ctx, storages.FileParams{FilePath: tempFile.Name(), StoreInterval: 0})
	assert.NoError(t, err)
	defer c.Close()

	_, err = c.GetMetric(ctx, metrics.TypeMetricCounter, "counter")
	assert.Error(t, err)

	var a int64

	for i := 0; i < Count; i++ {
		v := rand.Int64N(1024)
		a += v

		assert.NoError(t, c.UpdateMetric(ctx, metrics.TypeMetricCounter, "counter", v))

		val, err := c.GetMetric(ctx, metrics.TypeMetricCounter, "counter")
		assert.NoError(t, err)
		assert.Equal(t, a, val.(int64))
	}

	expected, err := storages.NewFile(ctx, storages.FileParams{
		FilePath:      tempFile.Name(),
		Restore:       true,
		StoreInterval: 0,
	})
	assert.NoError(t, err)
	defer expected.Close()

	assert.ElementsMatch(t, expected.GetAllMetrics(ctx), c.GetAllMetrics(ctx))

	_, err = c.GetMetric(ctx, metrics.TypeMetricCounter, "unknown")
	assert.Error(t, err)

	assert.Error(t, c.UpdateMetric(ctx, metrics.TypeMetricCounter, "counter", 1.0))
}

func TestFile_UpdateSequenceGauge(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempFile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)

	defer os.Remove(tempFile.Name())

	c, err := storages.NewFile(ctx, storages.FileParams{FilePath: tempFile.Name(), StoreInterval: 0})
	assert.NoError(t, err)
	defer c.Close()

	_, err = c.GetMetric(ctx, metrics.TypeMetricGauge, "gauge")
	assert.Error(t, err)

	for i := 0; i < Count; i++ {
		v := rand.Float64()

		assert.NoError(t, c.UpdateMetric(ctx, metrics.TypeMetricGauge, "gauge", v))

		val, err := c.GetMetric(ctx, metrics.TypeMetricGauge, "gauge")
		assert.NoError(t, err)
		assert.Equal(t, v, val.(float64))
	}

	expected, err := storages.NewFile(ctx, storages.FileParams{
		FilePath:      tempFile.Name(),
		Restore:       true,
		StoreInterval: 0,
	})
	assert.NoError(t, err)
	defer expected.Close()

	assert.ElementsMatch(t, expected.GetAllMetrics(ctx), c.GetAllMetrics(ctx))

	_, err = c.GetMetric(ctx, metrics.TypeMetricGauge, "unknown")
	assert.Error(t, err)

	assert.Error(t, c.UpdateMetric(ctx, metrics.TypeMetricGauge, "gauge", 1))
}

func TestFile_UpdateMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempFile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)

	defer os.Remove(tempFile.Name())

	c, err := storages.NewFile(ctx, storages.FileParams{FilePath: tempFile.Name(), StoreInterval: 0})
	assert.NoError(t, err)
	defer c.Close()

	sliceMetrics := []metrics.Metric{
		metrics.NewCounter("counter_1"),
		metrics.NewCounter("counter_2"),
		metrics.NewGauge("gauge_1"),
	}

	c.UpdateMetrics(ctx, sliceMetrics)

	expected, err := storages.NewFile(ctx, storages.FileParams{
		FilePath:      tempFile.Name(),
		Restore:       true,
		StoreInterval: 0,
	})
	assert.NoError(t, err)
	defer expected.Close()

	assert.ElementsMatch(t, expected.GetAllMetrics(ctx), c.GetAllMetrics(ctx))
	assert.ElementsMatch(t, sliceMetrics, c.GetAllMetrics(ctx))
}

func TestFile_SingleCounter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempFile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)

	defer os.Remove(tempFile.Name())

	c, err := storages.NewFile(ctx, storages.FileParams{FilePath: tempFile.Name(), StoreInterval: 1})
	assert.NoError(t, err)
	defer c.Close()

	var wg sync.WaitGroup

	wg.Add(NConsumers + NProducers)

	for i := 0; i < NProducers; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < Attempts; j++ {
				c.UpdateMetric(ctx, metrics.TypeMetricCounter, "counter", int64(1))
			}
		}()
	}

	for i := 0; i < NConsumers; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < Attempts; j++ {
				c.GetAllMetrics(ctx)
			}
		}()
	}

	wg.Wait()

	val, err := c.GetMetric(ctx, metrics.TypeMetricCounter, "counter")

	assert.Nil(t, err)
	assert.Equal(t, int64(NProducers*Attempts), val.(int64))

	time.Sleep(3 * time.Second)

	expected, err := storages.NewFile(ctx, storages.FileParams{
		FilePath:      tempFile.Name(),
		Restore:       true,
		StoreInterval: 0,
	})
	assert.NoError(t, err)
	defer expected.Close()

	assert.ElementsMatch(t, expected.GetAllMetrics(ctx), c.GetAllMetrics(ctx))
}

func TestFile_SeveralTypes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempFile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)

	defer os.Remove(tempFile.Name())

	c, err := storages.NewFile(ctx, storages.FileParams{FilePath: tempFile.Name(), StoreInterval: 1})
	assert.NoError(t, err)
	defer c.Close()

	var wg sync.WaitGroup

	wg.Add(NConsumers + NProducers)

	for i := 0; i < NProducers; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < Attempts; j++ {
				if j%2 == 0 {
					c.UpdateMetric(ctx, metrics.TypeMetricCounter, fmt.Sprintf("counter_%d", j%5), int64(1))
				} else {
					c.UpdateMetric(ctx, metrics.TypeMetricGauge, fmt.Sprintf("gauge_%d", j%5), rand.Float64())
				}
			}
		}()
	}

	for i := 0; i < NConsumers; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < Attempts; j++ {
				c.GetAllMetrics(ctx)
			}
		}()
	}

	wg.Wait()

	time.Sleep(3 * time.Second)

	expected, err := storages.NewFile(ctx, storages.FileParams{
		FilePath:      tempFile.Name(),
		Restore:       true,
		StoreInterval: 0,
	})
	assert.NoError(t, err)
	defer expected.Close()

	assert.ElementsMatch(t, expected.GetAllMetrics(ctx), c.GetAllMetrics(ctx))
}
