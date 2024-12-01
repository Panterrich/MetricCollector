package storages

import (
	"context"
	"database/sql"
	"errors"
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
		"CREATE TABLE IF NOT EXISTS metriccollector ("+
			"\"id\" VARCHAR(250) PRIMARY KEY, "+
			"\"type\" TEXT, "+
			"\"delta\" INTEGER, "+
			"\"value\" DOUBLE PRECISION"+
			");")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("exec context: %w", err)
	}

	return &Database{
		lock: sync.RWMutex{},
		DB:   db,
	}, nil
}

func (d *Database) getMetric(ctx context.Context, kind, name string) (any, error) {
	row := d.DB.QueryRowContext(ctx,
		"SELECT id, type, delta, value FROM metriccollector "+
			"WHERE id = $1 AND type = $2 LIMIT 1", name, kind)

	var (
		id    string
		mType string
		delta sql.NullInt64
		val   sql.NullFloat64
	)

	err := row.Scan(&id, &mType, &delta, &val)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, collector.ErrMetricNotFound
	} else if err != nil {
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

func (d *Database) GetMetric(ctx context.Context, kind, name string) (any, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return d.getMetric(ctx, kind, name)
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

	if rows.Err() != nil {
		return nil
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

	found := true

	oldValue, err := d.getMetric(ctx, kind, name)
	if errors.Is(err, collector.ErrMetricNotFound) {
		found = false
	} else if err != nil {
		return fmt.Errorf("get metric: %w", err)
	}

	jsonMetric := serialization.Metric{
		ID:    name,
		MType: kind,
	}

	if found {
		err = jsonMetric.SetValue(oldValue)
		if err != nil {
			return fmt.Errorf("metric set value: %w", err)
		}
	}

	metric, err := serialization.ConvertToMetric(jsonMetric)
	if err != nil {
		return fmt.Errorf("convert to metric: %w", err)
	}

	if ok := metric.Update(value); !ok {
		return collector.ErrUpdateMetric
	}

	if err = jsonMetric.SetValue(metric.Value()); err != nil {
		return fmt.Errorf("metric set new value: %w", err)
	}

	_, err = d.DB.ExecContext(ctx,
		"INSERT INTO metriccollector (id, type, delta, value) "+
			"VALUES ($1, $2, $3, $4) "+
			"ON CONFLICT(id) "+
			"DO UPDATE SET "+
			"delta = EXCLUDED.delta, "+
			"value = EXCLUDED.value;",
		jsonMetric.ID, jsonMetric.MType, jsonMetric.Delta, jsonMetric.Val)
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
		"INSERT INTO metriccollector (id, type, delta, value) "+
			"VALUES ($1, $2, $3, $4) "+
			"ON CONFLICT(id) "+
			"DO UPDATE SET "+
			"delta = EXCLUDED.delta, "+
			"value = EXCLUDED.value;")
	if err != nil {
		return fmt.Errorf("database prepare context: %w", err)
	}
	defer stmt.Close()

	for _, metric := range m {
		var (
			oldValue  any
			newMetric metrics.Metric
			found     = true
		)

		oldValue, err = d.getMetric(ctx, metric.Type(), metric.Name())
		if errors.Is(err, collector.ErrMetricNotFound) {
			found = false
		} else if err != nil {
			tx.Rollback()
			return fmt.Errorf("get metric: %w", err)
		}

		jsonMetric := serialization.Metric{
			ID:    metric.Name(),
			MType: metric.Type(),
		}

		if found {
			err = jsonMetric.SetValue(oldValue)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("metric set value: %w", err)
			}
		}

		newMetric, err = serialization.ConvertToMetric(jsonMetric)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("convert to metric: %w", err)
		}

		if ok := newMetric.Update(metric.Value()); !ok {
			tx.Rollback()
			return collector.ErrUpdateMetric
		}

		if err = jsonMetric.SetValue(newMetric.Value()); err != nil {
			tx.Rollback()
			return fmt.Errorf("metric set new value: %w", err)
		}

		_, err = stmt.ExecContext(ctx, jsonMetric.ID, jsonMetric.MType, jsonMetric.Delta, jsonMetric.Val)
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
