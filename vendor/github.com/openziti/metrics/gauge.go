package metrics

import (
	"github.com/rcrowley/go-metrics"
)

// Gauge represents a metric which is measuring a count and a rate
type Gauge interface {
	metrics.Gauge
	Metric
}

type gaugeImpl struct {
	metrics.Gauge
	dispose func()
}

func (gauge *gaugeImpl) Dispose() {
	gauge.dispose()
}
