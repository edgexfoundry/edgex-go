package metrics

import (
	"github.com/rcrowley/go-metrics"
	"time"
)

type Timer interface {
	Metric
	Time(func())
	Update(time.Duration)
	UpdateSince(time.Time)
}

type timerImpl struct {
	metrics.Timer
	dispose func()
}

func (t *timerImpl) Dispose() {
	t.Stop()
	t.dispose()
}
