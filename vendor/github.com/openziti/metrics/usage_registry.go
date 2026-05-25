package metrics

import (
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/metrics/metrics_pb"
	cmap "github.com/orcaman/concurrent-map/v2"
	"reflect"
	"sort"
	"sync"
	"time"
)

const (
	DefaultIntervalAgeThreshold = 80 * time.Second

	MinEventQueueSize     = 16
	DefaultEventQueueSize = 256
)

// Handler represents a sink for metric events
type Handler interface {
	// AcceptMetrics is called when new metrics become available
	AcceptMetrics(message *metrics_pb.MetricsMessage)
}

// UsageRegistry extends registry to allow collecting usage metrics
type UsageRegistry interface {
	Registry
	PollWithoutUsageMetrics() *metrics_pb.MetricsMessage
	IntervalCounter(name string, intervalSize time.Duration) IntervalCounter
	UsageCounter(name string, intervalSize time.Duration) UsageCounter
	FlushToHandler(handler Handler)
	StartReporting(eventSink Handler, reportInterval time.Duration, msgQueueSize int)
}

type UsageRegistryConfig struct {
	SourceId             string
	Tags                 map[string]string
	EventQueueSize       int
	CloseNotify          <-chan struct{}
	IntervalAgeThreshold time.Duration
}

func DefaultUsageRegistryConfig(sourceId string, closeNotify <-chan struct{}) UsageRegistryConfig {
	return UsageRegistryConfig{
		SourceId:             sourceId,
		Tags:                 map[string]string{},
		CloseNotify:          closeNotify,
		EventQueueSize:       DefaultEventQueueSize,
		IntervalAgeThreshold: DefaultIntervalAgeThreshold,
	}
}

func NewUsageRegistry(config UsageRegistryConfig) UsageRegistry {
	if config.EventQueueSize < MinEventQueueSize {
		config.EventQueueSize = MinEventQueueSize
	}

	if config.IntervalAgeThreshold <= 0 {
		config.IntervalAgeThreshold = DefaultIntervalAgeThreshold
	}

	registry := &usageRegistryImpl{
		registryImpl: registryImpl{
			sourceId:  config.SourceId,
			tags:      config.Tags,
			metricMap: cmap.New[Metric](),
		},
		intervalMetrics: cmap.New[intervalMetric](),
		eventChan:       make(chan func(), config.EventQueueSize),
		closeNotify:     config.CloseNotify,
	}

	return registry
}

type bucketEvent struct {
	interval *metrics_pb.MetricsMessage_IntervalCounter
	name     string
}

type intervalMetric interface {
	flushIntervals()
}

type usageRegistryImpl struct {
	registryImpl
	intervalMetrics      cmap.ConcurrentMap[string, intervalMetric]
	eventChan            chan func()
	intervalBuckets      []*bucketEvent
	usageBuckets         []*metrics_pb.MetricsMessage_UsageCounter
	closeNotify          <-chan struct{}
	lock                 sync.Mutex
	intervalAgeThreshold time.Duration
}

func (self *usageRegistryImpl) StartReporting(eventSink Handler, reportInterval time.Duration, msgQueueSize int) {
	msgEvents := make(chan *messageBuilder, msgQueueSize)
	go self.run(reportInterval, msgEvents)
	go self.sendMsgs(eventSink, msgEvents)
}

// IntervalCounter creates an IntervalCounter
func (self *usageRegistryImpl) IntervalCounter(name string, intervalSize time.Duration) IntervalCounter {
	self.lock.Lock()
	defer self.lock.Unlock()

	metric, present := self.metricMap.Get(name)
	if present {
		intervalCounter, ok := metric.(IntervalCounter)
		if !ok {
			panic(fmt.Errorf("metric '%v' already exists and is not an interval counter. It is a %v", name, reflect.TypeOf(metric).Name()))
		}
		return intervalCounter
	}

	disposeF := func() {
		self.dispose(name)
		self.intervalMetrics.Remove(name)
	}

	intervalCounter := newIntervalCounter(name, intervalSize, self.intervalAgeThreshold, self.eventChan, self, disposeF)
	self.metricMap.Set(name, intervalCounter)
	self.intervalMetrics.Set(name, intervalCounter)
	return intervalCounter
}

// UsageCounter creates a UsageCounter
func (self *usageRegistryImpl) UsageCounter(name string, intervalSize time.Duration) UsageCounter {
	self.lock.Lock()
	defer self.lock.Unlock()

	metric, present := self.metricMap.Get(name)
	if present {
		counter, ok := metric.(UsageCounter)
		if !ok {
			panic(fmt.Errorf("metric '%v' already exists and is not a usage counter. It is a %v", name, reflect.TypeOf(metric).Name()))
		}
		return counter
	}

	disposeF := func() { self.dispose(name) }
	usageCounter := newUsageCounter(name, intervalSize, self.intervalAgeThreshold, self, disposeF, self.eventChan)
	self.metricMap.Set(name, usageCounter)
	self.intervalMetrics.Set(name, usageCounter)
	return usageCounter
}

func (self *usageRegistryImpl) Poll() *metrics_pb.MetricsMessage {
	base := self.registryImpl.Poll()
	if base == nil && self.intervalBuckets == nil {
		return nil
	}

	var builder *messageBuilder
	if base == nil {
		builder = newMessageBuilder(self.sourceId, self.tags)
	} else {
		builder = (*messageBuilder)(base)
	}

	builder.addIntervalBucketEvents(self.intervalBuckets)
	self.intervalBuckets = nil

	builder.UsageCounters = self.usageBuckets
	self.usageBuckets = nil

	sort.Slice(builder.UsageCounters, func(i, j int) bool {
		return builder.UsageCounters[i].IntervalStartUTC < builder.UsageCounters[j].IntervalStartUTC
	})

	return (*metrics_pb.MetricsMessage)(builder)
}

func (self *usageRegistryImpl) PollWithoutUsageMetrics() *metrics_pb.MetricsMessage {
	return self.registryImpl.Poll()
}

func (self *usageRegistryImpl) reportInterval(counter *intervalCounterImpl, intervalStartUTC int64, values map[string]uint64) {
	bucket := &metrics_pb.MetricsMessage_IntervalBucket{
		IntervalStartUTC: intervalStartUTC,
		Values:           values,
	}

	interval := &metrics_pb.MetricsMessage_IntervalCounter{
		IntervalLength: uint64(counter.intervalSize.Seconds()),
		Buckets:        []*metrics_pb.MetricsMessage_IntervalBucket{bucket},
	}

	self.intervalBuckets = append(self.intervalBuckets, &bucketEvent{
		interval: interval,
		name:     counter.name,
	})
}

func (self *usageRegistryImpl) reportUsage(intervalStartUTC int64, intervalLength time.Duration, values map[string]*usageSet) {
	counter := &metrics_pb.MetricsMessage_UsageCounter{
		IntervalStartUTC: intervalStartUTC,
		IntervalLength:   uint64(intervalLength.Seconds()),
		Buckets:          map[string]*metrics_pb.MetricsMessage_UsageBucket{},
	}

	for k, v := range values {
		counter.Buckets[k] = &metrics_pb.MetricsMessage_UsageBucket{
			Values: v.values,
			Tags:   v.tags,
		}
	}
	self.usageBuckets = append(self.usageBuckets, counter)
}

func (self *usageRegistryImpl) run(reportInterval time.Duration, msgEvents chan *messageBuilder) {
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()

	for {
		select {
		case event := <-self.eventChan:
			event()
		case <-ticker.C:
			select {
			case msgEvents <- self.flushAndPoll():
			case <-self.closeNotify:
				return
			}
		case <-self.closeNotify:
			self.DisposeAll()
			return
		}
	}
}

func (self *usageRegistryImpl) flushAndPoll() *messageBuilder {
	for entry := range self.intervalMetrics.IterBuffered() {
		entry.Val.flushIntervals()
	}

	builder := newMessageBuilder(self.sourceId, self.tags)

	if len(self.intervalBuckets) > 0 {
		builder.addIntervalBucketEvents(self.intervalBuckets)
		self.intervalBuckets = nil
	}

	if len(self.usageBuckets) > 0 {
		builder.UsageCounters = self.usageBuckets
		self.usageBuckets = nil

		sort.Slice(builder.UsageCounters, func(i, j int) bool {
			return builder.UsageCounters[i].IntervalStartUTC < builder.UsageCounters[j].IntervalStartUTC
		})
	}

	return builder
}

func (registry *usageRegistryImpl) pollAppend(builder *messageBuilder) *metrics_pb.MetricsMessage {
	if len(builder.IntervalCounters) == 0 && len(builder.UsageCounters) == 0 && registry.metricMap.Count() == 0 {
		return nil
	}

	registry.EachMetric(func(name string, i Metric) {
		switch metric := i.(type) {
		case *gaugeImpl:
			builder.addIntGauge(name, metric.Snapshot())
		case *gaugeFloat64Impl:
			builder.addFloatGauge(name, metric.Snapshot())
		case *meterImpl:
			builder.addMeter(name, metric.Snapshot())
		case *histogramImpl:
			builder.addHistogram(name, metric.Snapshot())
		case *timerImpl:
			builder.addTimer(name, metric.Snapshot())
		case *intervalCounterImpl:
		// ignore, handled below
		case *usageCounterImpl:
			// ignore, handled below
		default:
			pfxlog.Logger().Errorf("Unsupported metric type %v", reflect.TypeOf(i))
		}
	})

	return (*metrics_pb.MetricsMessage)(builder)
}

func (self *usageRegistryImpl) sendMsgs(eventSink Handler, msgEvents chan *messageBuilder) {
	for {
		select {
		case builder := <-msgEvents:
			msg := self.pollAppend(builder)
			eventSink.AcceptMetrics(msg)
		case <-self.closeNotify:
			return
		}
	}
}

func (self *usageRegistryImpl) FlushAndPoll() *metrics_pb.MetricsMessage {
	msgC := make(chan *metrics_pb.MetricsMessage, 1)
	self.eventChan <- func() {
		builder := self.flushAndPoll()
		msg := self.pollAppend(builder)
		msgC <- msg
	}
	msg := <-msgC
	return msg
}

func (self *usageRegistryImpl) FlushToHandler(handler Handler) {
	if msg := self.FlushAndPoll(); msg != nil {
		handler.AcceptMetrics(msg)
	}
}
