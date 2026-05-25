package metrics

import (
	"github.com/rcrowley/go-metrics"
)

// Gauge represents a metric which is measuring a count and a rate
type Gauge interface {
	Metric
	Value() int64
	Update(int64)
}

type gaugeImpl struct {
	metrics.Gauge
	dispose func()
}

func (gauge *gaugeImpl) Dispose() {
	gauge.dispose()
}

// GaugeFloat64 represents a metric which holds a float64 value
type GaugeFloat64 interface {
	Metric
	Value() float64
	Update(float64)
}

type gaugeFloat64Impl struct {
	metrics.GaugeFloat64
	dispose func()
}

func (gauge *gaugeFloat64Impl) Dispose() {
	gauge.dispose()
}
