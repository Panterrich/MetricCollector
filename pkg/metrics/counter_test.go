package metrics_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Panterrich/MetricCollector/pkg/metrics"
)

func TestCounter(t *testing.T) {
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
			name:     "updated value #1",
			fields:   fields{name: "counter"},
			args:     []args{{int64(0), true}, {int64(2), true}, {int64(5), true}},
			expected: int64(7),
		},
		{
			name:     "updated value #2",
			fields:   fields{name: "counter"},
			args:     []args{{int64(-10), true}, {int64(0), true}, {int64(10), true}},
			expected: int64(0),
		},
		{
			name:     "invalid type, default value",
			fields:   fields{name: "counter"},
			args:     []args{{5.0, false}},
			expected: int64(0),
		},
		{
			name:     "invalid type, updated value",
			fields:   fields{name: "counter"},
			args:     []args{{0.0, false}, {int64(5), true}},
			expected: int64(5),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := metrics.NewCounter(tt.fields.name)

			for _, args := range tt.args {
				assert.Equal(t, args.want, c.Update(args.value))
			}

			assert.Equal(t, tt.fields.name, c.Name())
			assert.Equal(t, metrics.TypeMetricCounter, c.Type())
			assert.Equal(t, tt.expected, c.Value())

			nc := metrics.Clone(c)
			assert.Equal(t, nc, c)
		})
	}
}
