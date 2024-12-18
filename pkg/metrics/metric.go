package metrics

type Metric interface {
	Name() string
	Type() string
	Value() any
	Update(value any) bool
}

const (
	TypeMetricCounter = "counter"
	TypeMetricGauge   = "gauge"
)

var metricTypes = map[string]func(string) Metric{
	TypeMetricCounter: NewCounter,
	TypeMetricGauge:   NewGauge,
}

func NewMetric(kind, name string) Metric {
	if newMetric, ok := metricTypes[kind]; ok {
		return newMetric(name)
	}

	return nil
}

func Clone(m Metric) Metric {
	newMetric := NewMetric(m.Type(), m.Name())
	newMetric.Update(m.Value())

	return newMetric
}
