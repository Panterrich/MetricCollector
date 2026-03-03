package workpool_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Panterrich/MetricCollector/internal/storages"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
	"github.com/Panterrich/MetricCollector/pkg/workpool"
)

const (
	NConsumers = 10
	NProducers = 10
	Attempts   = 10000
)

func CaseWorkpool(t *testing.T, nWorkers int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := workpool.NewPool(ctx, nWorkers)
	m := storages.NewMemory()

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		for res := range pool.Results {
			assert.Equal(t, workpool.Result{}, res)
		}
	}()

	for i := 0; i < NConsumers; i++ {
		pool.Schedule(ctx, func(ctx context.Context) workpool.Result {
			for j := 0; j < Attempts; j++ {
				m.UpdateMetric(ctx, metrics.TypeMetricCounter, "counter", int64(1))
			}

			return workpool.Result{}
		})
	}

	for i := 0; i < NProducers; i++ {
		pool.Schedule(ctx, func(ctx context.Context) workpool.Result {
			for j := 0; j < Attempts; j++ {
				m.GetAllMetrics(ctx)
			}

			return workpool.Result{}
		})
	}

	cancel()
	pool.Wait()
	wg.Wait()
}

func TestWorkpool(t *testing.T) {
	type args struct {
		nWorkers int
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "single worker",
			args: args{nWorkers: 1},
		},
		{
			name: "double workers",
			args: args{nWorkers: 2},
		},
		{
			name: "multi workers",
			args: args{nWorkers: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CaseWorkpool(t, tt.args.nWorkers)
		})
	}
}
