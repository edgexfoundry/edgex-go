package metrics

import (
	"fmt"
	"sync/atomic"
	"time"
)

type MetricSink interface {
	Filter(name string) bool
	StartReport(registry Registry)
	EndReport(registry Registry)
	AcceptIntMetric(name string, value int64)
	AcceptFloatMetric(name string, value float64)
	AcceptPercentileMetric(name string, value PercentileSource)
}

type PercentileSource interface {
	Percentile(float64) float64
}

func NewDelegatingReporter(registry Registry, sink MetricSink, closeNotify <-chan struct{}) *DelegatingReporter {
	return &DelegatingReporter{
		registry:    registry,
		closeNotify: closeNotify,
		sink:        sink,
	}
}

type DelegatingReporter struct {
	registry    Registry
	closeNotify <-chan struct{}
	sink        MetricSink
	started     atomic.Bool
}

func (self *DelegatingReporter) Start(interval time.Duration) {
	if !self.started.CompareAndSwap(false, true) {
		return
	}

	timer := time.NewTicker(interval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			self.sink.StartReport(self.registry)
			self.registry.AcceptVisitor(self)
			self.sink.EndReport(self.registry)
		case <-self.closeNotify:
			return
		}
	}
}

func (self *DelegatingReporter) VisitIntMetric(name string, val int64, extra string) {
	if len(extra) > 0 {
		name = fmt.Sprintf("%s.%s", name, extra)
	}
	if self.sink.Filter(name) {
		self.sink.AcceptIntMetric(name, val)
	}
}

func (self *DelegatingReporter) VisitFloatMetric(name string, val float64, extra string) {
	if len(extra) > 0 {
		name = fmt.Sprintf("%s.%s", name, extra)
	}
	if self.sink.Filter(name) {
		self.sink.AcceptFloatMetric(name, val)
	}
}

func (self *DelegatingReporter) VisitPercentileMetric(name string, val PercentileSource, extra string) {
	if len(extra) > 0 {
		name = fmt.Sprintf("%s.%s", name, extra)
	}
	if self.sink.Filter(name) {
		self.sink.AcceptPercentileMetric(name, val)
	}
}

const (
	MetricNameCount      = "count"
	MetricNameMean       = "mean"
	MetricNameRateM1     = "rate_m1"
	MetricNameRateM5     = "rate_m5"
	MetricNameRateM15    = "rate_m15"
	MetricNameMin        = "min"
	MetricNameMax        = "max"
	MetricNamePercentile = "percentile"
)

func (self *DelegatingReporter) VisitGauge(name string, gauge Gauge) {
	self.VisitIntMetric(name, gauge.Value(), "")
}

func (self *DelegatingReporter) VisitGaugeFloat64(name string, gauge GaugeFloat64) {
	self.VisitFloatMetric(name, gauge.Value(), "")
}

func (self *DelegatingReporter) VisitMeter(name string, metric Meter) {
	self.VisitIntMetric(name, metric.Count(), MetricNameCount)
	self.VisitFloatMetric(name, metric.Rate1(), MetricNameRateM1)
	self.VisitFloatMetric(name, metric.Rate5(), MetricNameRateM5)
	self.VisitFloatMetric(name, metric.Rate15(), MetricNameRateM15)
	self.VisitFloatMetric(name, metric.RateMean(), MetricNameMean)
}

func (self *DelegatingReporter) VisitHistogram(name string, metric Histogram) {
	self.VisitIntMetric(name, metric.Count(), MetricNameCount)
	self.VisitFloatMetric(name, metric.Mean(), MetricNameMean)
	self.VisitIntMetric(name, metric.Min(), MetricNameMin)
	self.VisitIntMetric(name, metric.Max(), MetricNameMax)
	self.VisitPercentileMetric(name, metric, MetricNamePercentile)
}

func (self *DelegatingReporter) VisitTimer(name string, metric Timer) {
	self.VisitIntMetric(name, metric.Count(), MetricNameCount)

	self.VisitFloatMetric(name, metric.Rate1(), MetricNameRateM1)
	self.VisitFloatMetric(name, metric.Rate5(), MetricNameRateM5)
	self.VisitFloatMetric(name, metric.Rate15(), MetricNameRateM15)

	self.VisitFloatMetric(name, metric.Mean(), MetricNameMean)
	self.VisitIntMetric(name, metric.Min(), MetricNameMin)
	self.VisitIntMetric(name, metric.Max(), MetricNameMax)
	self.VisitPercentileMetric(name, metric, MetricNamePercentile)
}
