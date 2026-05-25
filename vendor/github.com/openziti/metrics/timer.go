package metrics

import (
	"github.com/rcrowley/go-metrics"
	"time"
)

type Timer interface {
	Metric
	Count() int64
	Max() int64
	Mean() float64
	Min() int64
	Percentile(float64) float64
	Percentiles([]float64) []float64
	Rate1() float64
	Rate5() float64
	Rate15() float64
	RateMean() float64
	StdDev() float64
	Sum() int64
	Variance() float64

	Time(func())
	Update(time.Duration)
	UpdateSince(time.Time)
	CreateSnapshot() Timer
}

type timerImpl struct {
	metrics.Timer
	dispose func()
}

func (t *timerImpl) CreateSnapshot() Timer {
	return &timerSnapshot{
		Timer: t.Snapshot(),
	}
}

func (t *timerImpl) Dispose() {
	t.Stop()
	t.dispose()
}

type timerSnapshot struct {
	metrics.Timer
}

func (t *timerSnapshot) Dispose() {
}

func (t *timerSnapshot) CreateSnapshot() Timer {
	return t
}
