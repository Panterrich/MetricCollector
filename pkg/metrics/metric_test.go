package metrics_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

func TestNewMetric_KnownTypes(t *testing.T) {
	type args struct {
		kind string
		name string
	}

	tests := []struct {
		name string
		args args
		want metrics.Metric
	}{
		{
			name: "counter",
			args: args{kind: metrics.TypeMetricCounter, name: "counter"},
			want: metrics.NewCounter("counter"),
		},
		{
			name: "gauge",
			args: args{kind: metrics.TypeMetricGauge, name: "gauge"},
			want: metrics.NewGauge("gauge"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if metric := metrics.NewMetric(tt.args.kind, tt.args.name); assert.NotNil(t, metric) {
				assert.Equal(t, tt.args.kind, metric.Type())
				assert.True(t, reflect.DeepEqual(metric, tt.want))
			}
		})
	}
}

func TestNewMetric_UnknownTypes(t *testing.T) {
	type args struct {
		kind string
		name string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "unknown type",
			args: args{kind: "unknown", name: "unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Nil(t, metrics.NewMetric(tt.args.kind, tt.args.name))
		})
	}
}
