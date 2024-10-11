package metrics

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetric_KnownTypes(t *testing.T) {
	type args struct {
		kind string
		name string
	}
	tests := []struct {
		name string
		args args
		want Metric
	}{
		{
			name: "counter",
			args: args{kind: TypeMetricCounter, name: "counter"},
			want: NewCounter("counter"),
		},
		{
			name: "gauge",
			args: args{kind: TypeMetricGauge, name: "gauge"},
			want: NewGauge("gauge"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if metric := NewMetric(tt.args.kind, tt.args.name); assert.NotNil(t, metric) {
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
			assert.Nil(t, NewMetric(tt.args.kind, tt.args.name))
		})
	}
}
