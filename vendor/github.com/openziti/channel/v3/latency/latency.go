/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package latency

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"sync/atomic"
	"time"
)

const (
	probeTime = 128
)

// LatencyHandler responds to latency messages with Result messages.
type LatencyHandler struct {
	responses int32
}

func (h *LatencyHandler) ContentType() int32 {
	return channel.ContentTypeLatencyType
}

func (h *LatencyHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	// need to send response in a separate go-routine. We get stuck sending, we'll also pause the receiving side
	// limit the number of concurrent responses
	if count := atomic.AddInt32(&h.responses, 1); count < 2 {
		go func() {
			defer atomic.AddInt32(&h.responses, -1)
			response := channel.NewResult(true, "")
			response.ReplyTo(msg)
			if err := response.WithPriority(channel.High).Send(ch); err != nil {
				pfxlog.ContextLogger(ch.Label()).WithError(err).Errorf("error sending latency response")
			}
		}()
	} else {
		atomic.AddInt32(&h.responses, -1)
	}
}

type ProbeConfig struct {
	Channel        channel.Channel
	Interval       time.Duration
	Timeout        time.Duration
	ResultHandler  func(resultNanos int64)
	TimeoutHandler func()
	ExitHandler    func()
}

func ProbeLatencyConfigurable(config *ProbeConfig) {
	ch := config.Channel
	log := pfxlog.ContextLogger(ch.Label())
	log.Debug("started")
	defer log.Debug("exited")
	defer func() {
		if config.ExitHandler != nil {
			config.ExitHandler()
		}
	}()

	for {
		time.Sleep(config.Interval)
		if ch.IsClosed() {
			return
		}

		request := channel.NewMessage(channel.ContentTypeLatencyType, nil)
		request.PutUint64Header(probeTime, uint64(time.Now().UnixNano()))
		response, err := request.WithPriority(channel.High).WithTimeout(config.Timeout).SendForReply(config.Channel)
		if err != nil {
			log.WithError(err).Error("unexpected error sending latency probe")
			if config.Channel.IsClosed() {
				log.WithError(err).Info("latency probe channel closed, exiting")
				return
			}
			if channel.IsTimeout(err) && config.TimeoutHandler != nil {
				config.TimeoutHandler()
			}
			continue
		}

		if sentTime, ok := response.GetUint64Header(probeTime); ok {
			latency := time.Now().UnixNano() - int64(sentTime)
			if config.ResultHandler != nil {
				config.ResultHandler(latency)
			}
		} else {
			log.Error("latency response did not contain probe time")
		}
	}
}

type Type uint8

const (
	RoundTripType  Type = 1
	BeforeSendType Type = 2
)

type Handler interface {
	HandleLatency(latencyType Type, latency time.Duration)
}

type HandlerF func(latencyType Type, latency time.Duration)

func (self HandlerF) HandleLatency(latencyType Type, latency time.Duration) {
	self(latencyType, latency)
}

func AddLatencyProbe(ch channel.Channel, binding channel.Binding, interval time.Duration, roundTripFreq uint8, handler HandlerF) {
	probe := &latencyProbe{
		handler:       handler,
		ch:            ch,
		interval:      interval,
		roundTripFreq: roundTripFreq,
	}
	binding.AddTypedReceiveHandler(probe)
	go probe.run()
}

type latencyProbe struct {
	handler       HandlerF
	ch            channel.Channel
	interval      time.Duration
	roundTripFreq uint8
	count         uint8
}

func (self *latencyProbe) ContentType() int32 {
	return channel.ContentTypeLatencyResponseType
}

func (self *latencyProbe) HandleReceive(m *channel.Message, _ channel.Channel) {
	if sentTime, ok := m.GetUint64Header(probeTime); ok {
		latency := time.Duration(time.Now().UnixNano() - int64(sentTime))
		self.handler(RoundTripType, latency)
	} else {
		pfxlog.Logger().Error("no send time on latency response")
	}
}

func (self *latencyProbe) run() {
	if self.ch.IsClosed() {
		pfxlog.ContextLogger(self.ch.Label()).Debug("exited")
		return
	}

	var sendable channel.Sendable
	if self.count == self.roundTripFreq {
		sendable = &RoundTripLatency{
			msg: channel.NewMessage(channel.ContentTypeLatencyType, nil),
		}
		self.count = 0
	} else {
		sendable = &SendTimeTracker{
			Handler:   self.handler,
			StartTime: time.Now(),
		}
	}
	if err := self.ch.Send(sendable); err != nil {
		pfxlog.ContextLogger(self.ch.Label()).WithError(err).Error("unexpected error sending latency probe")
		if self.ch.IsClosed() {
			return
		}
	}

	self.count++

	time.AfterFunc(self.interval, self.run)
}

type SendTimeTracker struct {
	channel.BaseSendable
	channel.BaseSendListener
	Handler   HandlerF
	StartTime time.Time
	seq       int32
}

func (self *SendTimeTracker) SetSequence(seq int32) {
	self.seq = seq
}

func (self *SendTimeTracker) Sequence() int32 {
	return self.seq
}

func (self *SendTimeTracker) SendListener() channel.SendListener {
	return self
}

func (self *SendTimeTracker) NotifyBeforeWrite() {
	t := time.Since(self.StartTime)
	self.Handler(BeforeSendType, t)
}

type RoundTripLatency struct {
	msg *channel.Message
	channel.BaseSendable
	channel.BaseSendListener
}

func (self *RoundTripLatency) SetSequence(seq int32) {
	self.msg.SetSequence(seq)
}

func (self *RoundTripLatency) Sequence() int32 {
	return self.msg.Sequence()
}

func (self *RoundTripLatency) Msg() *channel.Message {
	return self.msg
}

func (self *RoundTripLatency) SendListener() channel.SendListener {
	return self
}

func (self *RoundTripLatency) NotifyBeforeWrite() {
	self.msg.PutUint64Header(probeTime, uint64(time.Now().UnixNano()))
}

func AddLatencyProbeResponder(binding channel.Binding) {
	responder := &Responder{
		ch:              binding.GetChannel(),
		responseChannel: make(chan *channel.Message, 1),
	}
	binding.AddTypedReceiveHandler(responder)
	go responder.responseSender()
}

// Responder responds to latency messages with LatencyResponse messages.
type Responder struct {
	responseChannel chan *channel.Message
	ch              channel.Channel
}

func (self *Responder) ContentType() int32 {
	return channel.ContentTypeLatencyType
}

func (self *Responder) HandleReceive(msg *channel.Message, _ channel.Channel) {
	if sentTime, found := msg.Headers[probeTime]; found {
		resp := channel.NewMessage(channel.ContentTypeLatencyResponseType, nil)
		resp.Headers[probeTime] = sentTime
		select {
		case self.responseChannel <- resp:
		default:
		}
	}
}

func (self *Responder) responseSender() {
	log := pfxlog.ContextLogger(self.ch.Label())
	log.Debug("started")
	defer log.Debug("exited")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case response := <-self.responseChannel:
			if err := response.WithPriority(channel.High).Send(self.ch); err != nil {
				log.WithError(err).Error("error sending latency response")
				if self.ch.IsClosed() {
					return
				}
			}
		case <-ticker.C:
			if self.ch.IsClosed() {
				return
			}
		}
	}
}
