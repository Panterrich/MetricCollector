package storages_test

import (
	"context"
	"database/sql/driver"
	"math/rand/v2"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/storages"
	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

func TestDatabase_UpdateSequenceCounter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	mock.ExpectExec("CREATE TABLE").WillReturnResult(driver.ResultNoRows)

	c, err := storages.NewDatabase(ctx, storages.DatabaseParams{DB: db})
	assert.NoError(t, err)
	defer c.Close()

	_, err = c.GetMetric(ctx, metrics.TypeMetricCounter, "counter")
	assert.Error(t, err)

	var (
		a   int64
		val any
	)

	for i := 0; i < Count; i++ {
		v := rand.Int64N(1024)

		rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
			AddRow(1, "counter", a, nil)

		mock.ExpectQuery("SELECT id, type, delta, value").WillReturnRows(rows)
		mock.ExpectExec("INSERT INTO metriccollector").WillReturnResult(sqlmock.NewResult(1, 1))

		a += v

		assert.NoError(t, c.UpdateMetric(ctx, metrics.TypeMetricCounter, "counter", v))

		rows = sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
			AddRow(1, "counter", a, nil)

		mock.ExpectQuery("SELECT id, type, delta, value").WillReturnRows(rows)

		val, err = c.GetMetric(ctx, metrics.TypeMetricCounter, "counter")
		assert.NoError(t, err)
		assert.Equal(t, a, val.(int64))
	}

	mock.ExpectQuery("SELECT id, type, delta, value").WillReturnError(collector.ErrMetricNotFound)

	_, err = c.GetMetric(ctx, metrics.TypeMetricCounter, "unknown")
	assert.Error(t, err)

	assert.Error(t, c.UpdateMetric(ctx, metrics.TypeMetricCounter, "counter", 1.0))
}

func TestDatabase_UpdateSequenceGauge(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	mock.ExpectExec("CREATE TABLE").WillReturnResult(driver.ResultNoRows)

	c, err := storages.NewDatabase(ctx, storages.DatabaseParams{DB: db})
	assert.NoError(t, err)
	defer c.Close()

	_, err = c.GetMetric(ctx, metrics.TypeMetricGauge, "gauge")
	assert.Error(t, err)

	var val any

	for i := 0; i < Count; i++ {
		v := rand.Float64()

		rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
			AddRow(1, "gauge", nil, v)

		mock.ExpectQuery("SELECT id, type, delta, value").WillReturnRows(rows)
		mock.ExpectExec("INSERT INTO metriccollector").WillReturnResult(sqlmock.NewResult(1, 1))

		assert.NoError(t, c.UpdateMetric(ctx, metrics.TypeMetricGauge, "gauge", v))

		rows = sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
			AddRow(1, "gauge", nil, v)

		mock.ExpectQuery("SELECT id, type, delta, value").WillReturnRows(rows)

		val, err = c.GetMetric(ctx, metrics.TypeMetricGauge, "gauge")
		assert.NoError(t, err)
		assert.Equal(t, v, val.(float64))
	}

	mock.ExpectQuery("SELECT id, type, delta, value").WillReturnError(collector.ErrMetricNotFound)

	_, err = c.GetMetric(ctx, metrics.TypeMetricGauge, "unknown")
	assert.Error(t, err)

	assert.Error(t, c.UpdateMetric(ctx, metrics.TypeMetricGauge, "gauge", 1))
}

func TestDatabase_UpdateMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	mock.ExpectExec("CREATE TABLE").WillReturnResult(driver.ResultNoRows)

	c, err := storages.NewDatabase(ctx, storages.DatabaseParams{DB: db})
	assert.NoError(t, err)
	defer c.Close()

	sliceMetrics := []metrics.Metric{
		metrics.NewCounter("counter_1"),
		metrics.NewCounter("counter_2"),
		metrics.NewGauge("gauge_1"),
	}

	mock.ExpectBegin()

	p := mock.ExpectPrepare("INSERT INTO metriccollector")

	mock.ExpectQuery("SELECT id, type, delta, value FROM metriccollector").
		WithArgs("counter_1", "counter").
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
			AddRow("counter_1", "counter", int64(0), nil))

	p.ExpectExec().
		WithArgs("counter_1", "counter", int64(0), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("SELECT id, type, delta, value FROM metriccollector").
		WithArgs("counter_2", "counter").
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
			AddRow("counter_2", "counter", int64(0), nil))

	p.ExpectExec().
		WithArgs("counter_2", "counter", int64(0), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("SELECT id, type, delta, value FROM metriccollector").
		WithArgs("gauge_1", "gauge").
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
			AddRow("gauge_1", "gauge", nil, float64(0)))

	p.ExpectExec().
		WithArgs("gauge_1", "gauge", nil, float64(0)).
		WillReturnResult(sqlmock.NewResult(3, 1))

	mock.ExpectCommit()

	assert.NoError(t, c.UpdateMetrics(ctx, sliceMetrics))

	mock.ExpectQuery("SELECT id, type, delta, value FROM metriccollector").
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
			AddRow("counter_1", "counter", int64(0), nil).
			AddRow("counter_2", "counter", int64(0), nil).
			AddRow("gauge_1", "gauge", nil, float64(0)))

	assert.Equal(t, sliceMetrics, c.GetAllMetrics(ctx))
}
