package storages

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

type Database struct {
	lock sync.RWMutex
	DB   *sql.DB
}

var _ collector.Collector = (*Database)(nil)

type DatabaseParams struct {
	DatabaseDsn string
}

func NewDatabase(dp DatabaseParams) (collector.Collector, error) {
	db, err := sql.Open("pgx", dp.DatabaseDsn)
	if err != nil {
		return nil, fmt.Errorf("database create \"%s\": %w", dp.DatabaseDsn, err)
	}

	return &Database{
		lock: sync.RWMutex{},
		DB:   db,
	}, nil
}

func (d *Database) GetMetric(_ context.Context, _, _ string) (any, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return nil, nil
}

func (d *Database) GetAllMetrics(_ context.Context) []metrics.Metric {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return nil
}

func (d *Database) UpdateMetric(_ context.Context, _, _ string, _ any) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	return nil
}

func (d *Database) UpdateMetrics(_ context.Context, _ []metrics.Metric) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	return nil
}

func (d *Database) Close() {
	d.DB.Close()
}
