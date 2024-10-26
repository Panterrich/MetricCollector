package metrics_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Panterrich/MetricCollector/internal/metrics"
)

func TestGauge(t *testing.T) {
	type fields struct {
		name string
	}

	type args struct {
		value any
		want  bool
	}

	tests := []struct {
		name     string
		fields   fields
		args     []args
		expected any
	}{
		{
			name:     "updated value",
			fields:   fields{name: "gauge"},
			args:     []args{{0.0, true}, {0.1, true}, {5.0, true}},
			expected: float64(5.0),
		},
		{
			name:     "invalid type, default value",
			fields:   fields{name: "gauge"},
			args:     []args{{5, false}},
			expected: float64(0.0),
		},
		{
			name:     "invalid type, updated value",
			fields:   fields{name: "gauge"},
			args:     []args{{5.0, true}, {1, false}},
			expected: float64(5.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := metrics.NewGauge(tt.fields.name)

			for _, args := range tt.args {
				assert.Equal(t, args.want, g.Update(args.value))
			}

			assert.Equal(t, tt.expected, g.Value())
			assert.Equal(t, tt.fields.name, g.Name())
			assert.Equal(t, metrics.TypeMetricGauge, g.Type())
		})
	}
}
