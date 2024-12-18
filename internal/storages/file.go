package storages

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
	"github.com/Panterrich/MetricCollector/pkg/serialization"
)

type File struct {
	lock     sync.Mutex
	filePath string

	ticker *time.Ticker
	stop   chan struct{}
	wg     sync.WaitGroup

	storage collector.Collector
}

var _ collector.Collector = (*File)(nil)

type FileParams struct {
	FilePath      string
	Restore       bool
	StoreInterval uint
}

func NewFile(ctx context.Context, fp FileParams) (collector.Collector, error) {
	fs := &File{
		lock:     sync.Mutex{},
		filePath: fp.FilePath,
		ticker:   nil,
		stop:     make(chan struct{}),
		wg:       sync.WaitGroup{},
		storage:  NewMemory(),
	}

	if fp.StoreInterval != 0 {
		fs.ticker = time.NewTicker(time.Duration(fp.StoreInterval) * time.Second)
		fs.wg.Add(1)

		go func() {
			defer fs.wg.Done()

			for {
				select {
				case <-fs.ticker.C:
					if err := serialization.Save(ctx, fs.storage, fs.filePath); err != nil {
						log.Error().Msgf("can't save database: %v", err)
					}
				case <-fs.stop:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	if fp.Restore {
		if err := serialization.Load(ctx, fs.storage, fs.filePath); err != nil {
			return nil, fmt.Errorf("file storage create: %w", err)
		}
	}

	return fs, nil
}

func (f *File) GetMetric(ctx context.Context, kind, name string) (any, error) {
	val, err := f.storage.GetMetric(ctx, kind, name)
	if err != nil {
		return nil, fmt.Errorf("file storage get metric: %w", err)
	}

	return val, nil
}

func (f *File) GetAllMetrics(ctx context.Context) []metrics.Metric {
	return f.storage.GetAllMetrics(ctx)
}

func (f *File) UpdateMetric(ctx context.Context, kind, name string, value any) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	if err := f.storage.UpdateMetric(ctx, kind, name, value); err != nil {
		return fmt.Errorf("file storage update metric: %w", err)
	}

	if f.ticker == nil {
		if err := serialization.Save(ctx, f.storage, f.filePath); err != nil {
			return fmt.Errorf("file storage save: %w", err)
		}
	}

	return nil
}

func (f *File) UpdateMetrics(ctx context.Context, metrics []metrics.Metric) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	for _, metric := range metrics {
		if err := f.storage.UpdateMetric(ctx, metric.Type(), metric.Name(), metric.Value()); err != nil {
			return fmt.Errorf("file storage update metric: %w", err)
		}

		if ctx.Err() != nil {
			return nil
		}
	}

	if f.ticker == nil {
		if err := serialization.Save(ctx, f.storage, f.filePath); err != nil {
			return fmt.Errorf("file storage save: %w", err)
		}
	}

	return nil
}

func (f *File) Close() {
	if f.ticker != nil {
		f.stop <- struct{}{}
	}

	f.wg.Wait()
}
