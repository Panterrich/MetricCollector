package runtimestats_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Panterrich/MetricCollector/internal/storages"
	runtimestats "github.com/Panterrich/MetricCollector/pkg/runtime-stats"
	"github.com/Panterrich/MetricCollector/pkg/workpool"
	"github.com/stretchr/testify/assert"
)

const (
	nWorkers = 4
)

func TestUpdateAllMetrics_AllUpdate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := workpool.NewPool(ctx, nWorkers)
	m := storages.NewMemory()

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case res := <-pool.Results:
				assert.Nil(t, res.Err)
			case <-ctx.Done():
				return
			}
		}
	}()

	runtimestats.UpdateAllMetrics(ctx, pool, m)

	cancel()
	pool.Wait()
	wg.Wait()
}

func TestUpdateAllMetrics_CtxCancel(_ *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	pool := workpool.NewPool(ctx, nWorkers)
	m := storages.NewMemory()

	runtimestats.UpdateAllMetrics(ctx, pool, m)

	pool.Wait()
}
