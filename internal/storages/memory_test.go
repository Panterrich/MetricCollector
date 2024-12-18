package storages_test

import (
	"context"
	"sync"
	"testing"

	"github.com/Panterrich/MetricCollector/internal/storages"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

const (
	NConsumers = 10
	NProducers = 10
	Attempts   = 10000
)

func TestMemory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := storages.NewMemory()

	var wg sync.WaitGroup

	wg.Add(NConsumers + NProducers)

	for i := 0; i < NConsumers; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < Attempts; j++ {
				m.UpdateMetric(ctx, metrics.TypeMetricCounter, "counter", int64(1))
			}
		}()
	}

	for i := 0; i < NProducers; i++ {
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
	assert.Equal(t, int64(NConsumers*Attempts), val.(int64))
}
