package storages

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
	"github.com/Panterrich/MetricCollector/pkg/serialization"
)

type Database struct {
	lock sync.RWMutex
	DB   *sql.DB
}

var _ collector.Collector = (*Database)(nil)

type DatabaseParams struct {
	DatabaseDsn string
}

func NewDatabase(ctx context.Context, dp DatabaseParams) (collector.Collector, error) {
	db, err := sql.Open("pgx", dp.DatabaseDsn)
	if err != nil {
		return nil, fmt.Errorf("database create \"%s\": %w", dp.DatabaseDsn, err)
	}

	_, err = db.ExecContext(ctx,
		"CREATE TABLE metriccollector ( "+
			"\"id\" VARCHAR(250) PRIMARY KEY, "+
			"\"type\" TEXT, "+
			"\"delta\" INTEGER, "+
			"\"value\" DOUBLE PRECISION "+
			")")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("exec context: %w", err)
	}

	return &Database{
		lock: sync.RWMutex{},
		DB:   db,
	}, nil
}

func (d *Database) GetMetric(ctx context.Context, kind, name string) (any, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	row := d.DB.QueryRowContext(ctx,
		"SELECT id, type, delta, value FROM metriccollector "+
			"WHERE id = ? AND type = ? LIMIT 1", name, kind)

	var (
		id    string
		mType string
		delta sql.NullInt64
		val   sql.NullFloat64
	)

	err := row.Scan(&id, &mType, &delta, &val)
	if err != nil {
		return nil, fmt.Errorf("row scan error: %w", err)
	}

	switch {
	case delta.Valid:
		if mType != metrics.TypeMetricCounter {
			return nil, collector.ErrInvalidMetricType
		}

		return delta.Int64, nil
	case val.Valid:
		if mType != metrics.TypeMetricGauge {
			return nil, collector.ErrInvalidMetricType
		}

		return val.Float64, nil
	default:
		return nil, collector.ErrMetricNotFound
	}
}

func (d *Database) GetAllMetrics(ctx context.Context) []metrics.Metric {
	d.lock.RLock()
	defer d.lock.RUnlock()

	m := make([]serialization.Metric, 0)

	rows, err := d.DB.QueryContext(ctx, "SELECT id, type, delta, value FROM metriccollector")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var (
		id    string
		mType string
		delta sql.NullInt64
		val   sql.NullFloat64
	)

	for rows.Next() {
		err = rows.Scan(&id, &mType, &delta, &val)
		if err != nil {
			return nil
		}

		metric := serialization.Metric{
			ID:    id,
			MType: mType,
		}

		switch {
		case delta.Valid:
			if mType != metrics.TypeMetricCounter {
				return nil
			}

			metric.Delta = &delta.Int64
		case val.Valid:
			if mType != metrics.TypeMetricGauge {
				return nil
			}

			metric.Val = &val.Float64
		default:
			return nil
		}

		m = append(m, metric)
	}

	metrics, err := serialization.ConvertToMetrics(m)
	if err != nil {
		return nil
	}

	return metrics
}

func (d *Database) UpdateMetric(ctx context.Context, kind, name string, value any) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	metric := serialization.Metric{
		ID:    name,
		MType: kind,
	}

	err := metric.SetValue(value)
	if err != nil {
		return fmt.Errorf("metric set value: %w", err)
	}

	_, err = d.DB.ExecContext(ctx,
		"INSERT INTO metricccolector (id, type, delta, value) "+
			"VALUES (@id, @type, @delta, @value) "+
			"ON CONFLICT(id) "+
			"DO UPDATE SET "+
			"delta = EXCLUDED.delta, "+
			"value = EXCLUDED.value;",
		sql.Named("id", metric.ID),
		sql.Named("type", metric.MType),
		sql.Named("delta", metric.Delta),
		sql.Named("value", metric.Val))
	if err != nil {
		return fmt.Errorf("database insert into mettriccollector: %w", err)
	}

	return nil
}

func (d *Database) UpdateMetrics(ctx context.Context, m []metrics.Metric) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if len(m) == 0 {
		return nil
	}

	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("database transaction: %w", err)
	}

	stmt, err := d.DB.PrepareContext(ctx,
		"INSERT INTO metricccolector (id, type, delta, value) "+
			"VALUES (@id, @type, @delta, @value) "+
			"ON CONFLICT(id) "+
			"DO UPDATE SET "+
			"delta = EXCLUDED.delta, "+
			"value = EXCLUDED.value;")
	if err != nil {
		return fmt.Errorf("database prepare context: %w", err)
	}
	defer stmt.Close()

	for _, metric := range m {
		jsonMetric := serialization.Metric{
			ID:    metric.Name(),
			MType: metric.Type(),
		}

		err = jsonMetric.SetValue(metric.Value())
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("metric set value: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			sql.Named("id", jsonMetric.ID),
			sql.Named("type", jsonMetric.MType),
			sql.Named("delta", jsonMetric.Delta),
			sql.Named("value", jsonMetric.Val))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("stmt exec context: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit: %w", err)
	}

	return nil
}

func (d *Database) Close() {
	d.DB.Close()
}