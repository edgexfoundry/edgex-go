package channel

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync/atomic"
	"time"
)

const (
	DefaultHeartbeatSendInterval  = 10 * time.Second
	DefaultHeartbeatCheckInterval = time.Second
	DefaultHeartbeatTimeout       = 30 * time.Second
)

type HeartbeatOptions struct {
	SendInterval             time.Duration `json:"sendInterval"`
	CheckInterval            time.Duration `json:"checkInterval"`
	CloseUnresponsiveTimeout time.Duration `json:"closeUnresponsiveTimeout"`
	src                      map[interface{}]interface{}
}

func (self *HeartbeatOptions) GetDuration(name string) (*time.Duration, error) {
	if value, found := self.src[name]; found {
		if strVal, ok := value.(string); ok {
			if d, err := time.ParseDuration(strVal); err == nil {
				return &d, nil
			} else {
				return nil, errors.Wrapf(err, "invalid value for %v: %v", name, value)
			}
		} else {
			return nil, errors.Errorf("invalid (non-string) value for %v: %v", name, value)
		}
	}
	return nil, nil
}

func DefaultHeartbeatOptions() *HeartbeatOptions {
	return &HeartbeatOptions{
		SendInterval:             DefaultHeartbeatSendInterval,
		CheckInterval:            DefaultHeartbeatCheckInterval,
		CloseUnresponsiveTimeout: DefaultHeartbeatTimeout,
	}
}

func LoadHeartbeatOptions(data map[interface{}]interface{}) (*HeartbeatOptions, error) {
	options := DefaultHeartbeatOptions()
	options.src = data

	if value, err := options.GetDuration("sendInterval"); err != nil {
		return nil, err
	} else if value != nil {
		options.SendInterval = *value
	}

	if value, err := options.GetDuration("checkInterval"); err != nil {
		return nil, err
	} else if value != nil {
		options.CheckInterval = *value
	}

	if value, err := options.GetDuration("closeUnresponsiveTimeout"); err != nil {
		return nil, err
	} else if value != nil {
		options.CheckInterval = *value
	}

	return options, nil
}

// HeartbeatCallback provide an interface that is notified when various heartbeat events take place
type HeartbeatCallback interface {
	HeartbeatTx(ts int64)
	HeartbeatRx(ts int64)
	HeartbeatRespTx(ts int64)
	HeartbeatRespRx(ts int64)
	CheckHeartBeat()
}

// ConfigureHeartbeat setups up heartbeats on the given channel. It assumes that an equivalent setup happens on the
// other side of the channel.
//
// When possible, heartbeats will be sent on existing traffic. When a heartbeat is due to be sent, the next message sent
// will include a heartbeat header. If no message is sent by the time the checker runs on checkInterval, a standalone
// heartbeat message will be sent.
//
// Similarly, when a message with a heartbeat header is received, the next sent message will have a header set with
// the heartbeat response. If no message is sent within a few milliseconds, a standalone heartbeat response will be
// sent
func ConfigureHeartbeat(binding Binding, heartbeatInterval time.Duration, checkInterval time.Duration, cb HeartbeatCallback) {
	hb := &heartbeater{
		ch:                  binding.GetChannel(),
		heartBeatIntervalNs: heartbeatInterval.Nanoseconds(),
		callback:            cb,
		events:              make(chan heartbeatEvent, 4),
	}

	binding.AddReceiveHandler(ContentTypeHeartbeat, hb)
	binding.AddTransformHandler(hb)

	go hb.pulse(checkInterval)
}

// Note: if altering this struct, be sure to account for 64 bit alignment on 32 bit arm arch
// https://pkg.go.dev/sync/atomic#pkg-note-BUG
// https://github.com/golang/go/issues/36606
type heartbeater struct {
	ch                   Channel
	lastHeartbeatTx      int64
	heartBeatIntervalNs  int64
	unrespondedHeartbeat int64
	callback             HeartbeatCallback
	events               chan heartbeatEvent
}

func (self *heartbeater) HandleReceive(*Message, Channel) {
	// ignore incoming heartbeat events, everything is handled by the transformer
}

func (self *heartbeater) queueEvent(event heartbeatEvent) {
	select {
	case self.events <- event:
	default:
	}
}

func (self *heartbeater) Rx(m *Message, _ Channel) {
	if val, found := m.GetUint64Header(HeartbeatHeader); found {
		self.queueEvent(heartbeatRxEvent(val))
	}

	if val, found := m.GetUint64Header(HeartbeatResponseHeader); found {
		self.queueEvent(heartbeatRespRxEvent(val))
	}
}

func (self *heartbeater) Tx(m *Message, _ Channel) {
	if m.ContentType == ContentTypeRaw {
		return
	}
	now := time.Now().UnixNano()
	if now-self.lastHeartbeatTx > self.heartBeatIntervalNs {
		m.PutUint64Header(HeartbeatHeader, uint64(now))
		atomic.StoreInt64(&self.lastHeartbeatTx, now)
		self.queueEvent(heartbeatTxEvent(now))
	}

	if unrespondedHeartbeat := atomic.LoadInt64(&self.unrespondedHeartbeat); unrespondedHeartbeat != 0 {
		m.PutUint64Header(HeartbeatResponseHeader, uint64(unrespondedHeartbeat))
		atomic.StoreInt64(&self.unrespondedHeartbeat, 0)
		self.queueEvent(heartbeatRespTxEvent(now))
	}
}

func (self *heartbeater) pulse(checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for !self.ch.IsClosed() {
		select {
		case tick := <-ticker.C:
			now := tick.UnixNano()
			lastHeartbeatTx := atomic.LoadInt64(&self.lastHeartbeatTx)
			if now-lastHeartbeatTx > self.heartBeatIntervalNs {
				self.sendHeartbeat()
			}
			self.callback.CheckHeartBeat()

		case event := <-self.events:
			event.handle(self)
		}
	}
}

func (self *heartbeater) sendHeartbeat() {
	m := NewMessage(ContentTypeHeartbeat, nil) // don't need to add heartbeat
	if err := m.WithTimeout(time.Second).SendAndWaitForWire(self.ch); err != nil && !self.ch.IsClosed() {
		logrus.WithError(err).
			WithField("channelId", self.ch.Label()).
			Error("pulse failed to send heartbeat")
	}
}

func (self *heartbeater) sendHeartbeatIfQueueFree() {
	m := NewMessage(ContentTypeHeartbeat, nil) // don't need to add heartbeat
	if err := m.WithTimeout(10 * time.Millisecond).Send(self.ch); err != nil && !self.ch.IsClosed() {
		logrus.WithError(err).
			WithField("channelId", self.ch.Label()).
			Error("handleUnresponded failed to send heartbeat")
	}
}

type heartbeatEvent interface {
	handle(heartbeater *heartbeater)
}

type heartbeatTxEvent int64

func (h heartbeatTxEvent) handle(heartbeater *heartbeater) {
	heartbeater.callback.HeartbeatTx(int64(h))
}

type heartbeatRxEvent int64

func (h heartbeatRxEvent) handle(heartbeater *heartbeater) {
	atomic.StoreInt64(&heartbeater.unrespondedHeartbeat, int64(h))
	heartbeater.callback.HeartbeatRx(int64(h))

	// wait a few milliseconds to allowing already queued traffic to respond to the heartbeat
	time.AfterFunc(2*time.Millisecond, func() {
		select {
		case heartbeater.events <- handleUnresponded{}:
		default:
		}
	})
}

type heartbeatRespTxEvent int64

func (h heartbeatRespTxEvent) handle(heartbeater *heartbeater) {
	heartbeater.callback.HeartbeatRespTx(int64(h))
}

type heartbeatRespRxEvent int64

func (h heartbeatRespRxEvent) handle(heartbeater *heartbeater) {
	heartbeater.callback.HeartbeatRespRx(int64(h))
}

type handleUnresponded struct{}

func (h handleUnresponded) handle(heartbeater *heartbeater) {
	if unrespondedHeartbeat := atomic.LoadInt64(&heartbeater.unrespondedHeartbeat); unrespondedHeartbeat != 0 {
		heartbeater.sendHeartbeatIfQueueFree()
	}
}
