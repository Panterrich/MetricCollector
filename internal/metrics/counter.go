package metrics

type Counter struct {
	name  string
	value int64
}

var _ Metric = (*Counter)(nil)

func NewCounter(name string) Metric {
	return &Counter{
		name:  name,
	}
}

func (c *Counter) Name() string {
	return c.name
}

func (c *Counter) Type() string {
	return TypeMetricCounter
}

func (c *Counter) Value() any {
	return c.value
}

func (c *Counter) Update(value any) bool {
	if v, ok := value.(int64); ok {
		c.value += v
		return true
	}

	return false
}
