package metrics

type Gauge struct {
	name  string
	value float64
}

var _ Metric = (*Gauge)(nil)

func NewGauge(name string) Metric {
	return &Gauge{
		name:  name,
	}
}

func (g *Gauge) Name() string {
	return g.name
}

func (g *Gauge) Type() string {
	return TypeMetricGauge
}

func (g *Gauge) Value() any {
	return g.value
}

func (g *Gauge) Update(value any) bool {
	if v, ok := value.(float64); ok {
		g.value = v
		return true
	}

	return false
}
