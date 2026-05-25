package network

import (
	"fmt"
	"github.com/openziti/channel/v4"
	"github.com/openziti/metrics"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

type timingReceiveHandler struct {
	handler channel.ReceiveHandler
	timer   metrics.Histogram
}

func (t *timingReceiveHandler) HandleReceive(m *channel.Message, ch channel.Channel) {
	start := time.Now()
	t.handler.HandleReceive(m, ch)
	t.timer.Update(int64(time.Since(start)))
}

func NewMessageTimingBinding(binding channel.Binding) channel.Binding {
	registry := metrics.NewRegistry("", nil)
	wrapper := &messageTimingBinding{
		binding:  binding,
		registry: registry,
	}
	closeNotify := make(chan struct{})
	closed := atomic.Bool{}
	binding.AddCloseHandler(channel.CloseHandlerF(func(ch channel.Channel) {
		if closed.CompareAndSwap(false, true) {
			close(closeNotify)
		}
	}))
	reporter := metrics.NewDelegatingReporter(registry, wrapper, closeNotify)
	go reporter.Start(5 * time.Second)
	return wrapper
}

type messageTimingBinding struct {
	registry metrics.Registry
	binding  channel.Binding
	output   []string
}

func (self *messageTimingBinding) StartReport(metrics.Registry) {}

func (self *messageTimingBinding) EndReport(metrics.Registry) {
	sort.Strings(self.output)
	for _, output := range self.output {
		fmt.Println(output)
	}
	self.output = nil
}

func (self *messageTimingBinding) Printf(msg string, args ...interface{}) {
	self.output = append(self.output, fmt.Sprintf(msg, args...))
}

func (self *messageTimingBinding) Filter(name string) bool {
	return strings.HasSuffix(name, metrics.MetricNameCount) ||
		strings.HasSuffix(name, metrics.MetricNamePercentile)
}

func (self *messageTimingBinding) AcceptIntMetric(name string, value int64) {
	self.Printf("%s -> %d", name, value)
}

func (self *messageTimingBinding) AcceptFloatMetric(name string, value float64) {
	self.Printf("%s -> %f", name, value)
}

func (self *messageTimingBinding) AcceptPercentileMetric(name string, value metrics.PercentileSource) {
	self.Printf("%s.50p -> %s", name, time.Duration(value.Percentile(.5)).String())
	self.Printf("%s.75p -> %s", name, time.Duration(value.Percentile(.75)).String())
	self.Printf("%s.95p -> %s", name, time.Duration(value.Percentile(.95)).String())
}

func (self *messageTimingBinding) Bind(h channel.BindHandler) error {
	return h.BindChannel(self)
}

func (self *messageTimingBinding) AddPeekHandler(h channel.PeekHandler) {
	self.binding.AddPeekHandler(h)
}

func (self *messageTimingBinding) AddTransformHandler(h channel.TransformHandler) {
	self.binding.AddTransformHandler(h)
}

func (self *messageTimingBinding) AddReceiveHandler(contentType int32, h channel.ReceiveHandler) {
	timer := self.registry.Histogram(fmt.Sprintf("msgs.%d.time", contentType))
	self.binding.AddReceiveHandler(contentType, &timingReceiveHandler{
		handler: h,
		timer:   timer,
	})
}

func (self *messageTimingBinding) AddReceiveHandlerF(contentType int32, h channel.ReceiveHandlerF) {
	self.AddReceiveHandler(contentType, h)
}

func (self *messageTimingBinding) AddTypedReceiveHandler(h channel.TypedReceiveHandler) {
	self.AddReceiveHandler(h.ContentType(), h)
}

func (self *messageTimingBinding) AddErrorHandler(h channel.ErrorHandler) {
	self.binding.AddErrorHandler(h)
}

func (self *messageTimingBinding) AddCloseHandler(h channel.CloseHandler) {
	self.binding.AddCloseHandler(h)
}

func (self *messageTimingBinding) SetUserData(data interface{}) {
	self.binding.SetUserData(data)
}

func (self *messageTimingBinding) GetUserData() interface{} {
	return self.binding.GetUserData()
}

func (self *messageTimingBinding) GetChannel() channel.Channel {
	return self.binding.GetChannel()
}
