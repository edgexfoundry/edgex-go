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

package channel

import (
	"bytes"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/openziti/foundation/v2/info"
	"github.com/openziti/foundation/v2/sequence"
)

type MultiChannelConfig struct {
	LogicalName     string
	Options         *Options
	UnderlayHandler UnderlayHandler
	BindHandler     BindHandler
	Underlay        Underlay

	InjectUnderlayTypeIntoMessages bool
}

type senderContextImpl struct {
	sequence    *sequence.Sequence
	closeNotify chan struct{}
}

func (self *senderContextImpl) NextSequence() int32 {
	return int32(self.sequence.Next())
}

func (self *senderContextImpl) GetCloseNotify() chan struct{} {
	return self.closeNotify
}

func NewSenderContext() SenderContext {
	return &senderContextImpl{
		sequence:    sequence.NewSequence(),
		closeNotify: make(chan struct{}),
	}
}

type multiChannelImpl struct {
	// Note: if altering this struct, be sure to account for 64 bit alignment on 32 bit arm arch
	// https://pkg.go.dev/sync/atomic#pkg-note-BUG
	// https://github.com/golang/go/issues/36606
	lastRead int64

	ownerId          string
	channelId        string
	logicalName      string
	fallbackUnderlay atomic.Pointer[Underlay]

	options           *Options
	waiters           waiterMap
	flags             concurrenz.AtomicBitSet
	closeNotify       chan struct{}
	peekHandlers      []PeekHandler
	transformHandlers []TransformHandler
	receiveHandlers   map[int32]ReceiveHandler
	errorHandlers     []ErrorHandler
	closeHandlers     []CloseHandler
	underlayHandler   UnderlayHandler
	userData          interface{}
	replyCounter      uint32
	groupSecret       []byte

	lock      sync.Mutex
	underlays concurrenz.CopyOnWriteSlice[Underlay]
}

func NewMultiChannel(config *MultiChannelConfig) (MultiChannel, error) {
	if config.UnderlayHandler == nil {
		return nil, fmt.Errorf("no underlay handler configured for multi channel %s", config.LogicalName)
	}

	if config.Underlay == nil {
		return nil, errors.New("unable to initialize multi channel (initialization produced zero underlays)")
	}

	impl := &multiChannelImpl{
		channelId:       config.Underlay.ConnectionId(),
		logicalName:     config.LogicalName,
		options:         config.Options,
		receiveHandlers: map[int32]ReceiveHandler{},
		closeNotify:     config.UnderlayHandler.GetCloseNotify(),
		underlayHandler: config.UnderlayHandler,
	}

	impl.flags.Set(flagInjectUnderlayType, config.InjectUnderlayTypeIntoMessages)

	impl.ownerId = config.Underlay.Id()
	impl.fallbackUnderlay.Store(&config.Underlay)
	impl.underlays.Append(config.Underlay)

	groupSecret := config.Underlay.Headers()[GroupSecretHeader]
	if len(groupSecret) == 0 {
		return nil, errors.New("no group secret header found for multi channel")
	}
	impl.groupSecret = groupSecret

	config.UnderlayHandler.ChannelCreated(impl)

	if err := bind(config.BindHandler, impl); err != nil {
		for _, u := range impl.underlays.Value() {
			if closeErr := u.Close(); closeErr != nil {
				if !errors.Is(closeErr, net.ErrClosed) {
					pfxlog.ContextLogger(impl.Label()).WithError(err).Warn("error closing underlay")
				}
			}
		}
		return nil, err
	}

	impl.startMultiplex(config.Underlay)
	impl.underlayHandler.HandleUnderlayAccepted(impl, config.Underlay)
	go impl.underlayHandler.Start(impl)

	return impl, nil
}

func (self *multiChannelImpl) AcceptUnderlay(underlay Underlay) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	groupSecret := underlay.Headers()[GroupSecretHeader]
	if !bytes.Equal(groupSecret, self.groupSecret) {
		if err := underlay.Close(); err != nil {
			pfxlog.ContextLogger(self.Label()).WithError(err).Error("error closing underlay")
		}
		return fmt.Errorf("new underlay for '%s' not accepted: incorrect group secret", self.ConnectionId())
	}

	if self.IsClosed() {
		if err := underlay.Close(); err != nil {
			pfxlog.ContextLogger(self.Label()).WithError(err).Error("error closing underlay")
		}
		return fmt.Errorf("new underlay for '%s' not accepted: multi-channel is closed", self.ConnectionId())
	}

	self.fallbackUnderlay.Store(&underlay)
	self.underlays.Append(underlay)

	self.startMultiplex(underlay)

	self.underlayHandler.HandleUnderlayAccepted(self, underlay)

	return nil
}

func (self *multiChannelImpl) startMultiplex(underlay Underlay) {
	notifier := NewCloseNotifier()
	go self.Rxer(underlay, notifier)
	go self.Txer(underlay, notifier)
}

func (self *multiChannelImpl) GetUnderlayCountsByType() map[string]int {
	result := map[string]int{}
	for _, u := range self.underlays.Value() {
		underlayType := GetUnderlayType(u)
		result[underlayType]++
	}
	return result
}

func (self *multiChannelImpl) CloseNotify() <-chan struct{} {
	return self.closeNotify
}

func (self *multiChannelImpl) GetUnderlays() []Underlay {
	return slices.Clone(self.underlays.Value())
}

func (self *multiChannelImpl) Send(s Sendable) error {
	return self.underlayHandler.GetDefaultSender().Send(s)
}

func (self *multiChannelImpl) TrySend(s Sendable) (bool, error) {
	return self.underlayHandler.GetDefaultSender().TrySend(s)
}

func (self *multiChannelImpl) Id() string {
	return self.ownerId
}

func (self *multiChannelImpl) LogicalName() string {
	return self.logicalName
}

func (self *multiChannelImpl) SetLogicalName(logicalName string) {
	self.logicalName = logicalName
}

func (self *multiChannelImpl) ConnectionId() string {
	return self.channelId
}

func (self *multiChannelImpl) Certificates() []*x509.Certificate {
	return self.Underlay().Certificates()
}

func (self *multiChannelImpl) Headers() map[int32][]byte {
	return self.Underlay().Headers()
}

func (self *multiChannelImpl) Label() string {
	return fmt.Sprintf("ch{%s}->%s", self.LogicalName(), self.Underlay().Label())
}

func (self *multiChannelImpl) GetOptions() *Options {
	return self.options
}

func (self *multiChannelImpl) GetChannel() Channel {
	return self
}

func (self *multiChannelImpl) Bind(h BindHandler) error {
	return h.BindChannel(self)
}

func (self *multiChannelImpl) AddPeekHandler(h PeekHandler) {
	self.peekHandlers = append(self.peekHandlers, h)
}

func (self *multiChannelImpl) AddTransformHandler(h TransformHandler) {
	self.transformHandlers = append(self.transformHandlers, h)
}

func (self *multiChannelImpl) AddTypedReceiveHandler(h TypedReceiveHandler) {
	self.receiveHandlers[h.ContentType()] = h
}

func (self *multiChannelImpl) AddReceiveHandler(contentType int32, h ReceiveHandler) {
	self.receiveHandlers[contentType] = h
}

func (self *multiChannelImpl) AddReceiveHandlerF(contentType int32, h ReceiveHandlerF) {
	self.AddReceiveHandler(contentType, h)
}

func (self *multiChannelImpl) AddErrorHandler(h ErrorHandler) {
	self.errorHandlers = append(self.errorHandlers, h)
}

func (self *multiChannelImpl) AddCloseHandler(h CloseHandler) {
	self.closeHandlers = append(self.closeHandlers, h)
}

func (self *multiChannelImpl) SetUserData(data interface{}) {
	self.userData = data
}

func (self *multiChannelImpl) GetUserData() interface{} {
	return self.userData
}

func (self *multiChannelImpl) GetUnderlayHandler() UnderlayHandler {
	return self.underlayHandler
}

func (self *multiChannelImpl) Close() error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if self.flags.CompareAndSet(flagClosed, false, true) {
		pfxlog.ContextLogger(self.Label()).Debug("closing channel")

		close(self.closeNotify)

		for _, peekHandler := range self.peekHandlers {
			peekHandler.Close(self)
		}

		if len(self.closeHandlers) > 0 {
			for _, closeHandler := range self.closeHandlers {
				closeHandler.HandleClose(self)
			}
		} else {
			pfxlog.ContextLogger(self.Label()).Debug("no close handlers")
		}

		var errs []error
		for _, u := range self.underlays.Value() {
			if err := u.Close(); err != nil {
				errs = append(errs, err)
			}
		}

		return errors.Join(errs...)
	}

	return nil
}

func (self *multiChannelImpl) IsClosed() bool {
	return self.flags.IsSet(flagClosed)
}

func (self *multiChannelImpl) Underlay() Underlay {
	return *self.fallbackUnderlay.Load()
}

func (self *multiChannelImpl) Rx(m *Message) {
	log := pfxlog.ContextLogger(self.Label())

	now := info.NowInMilliseconds()
	atomic.StoreInt64(&self.lastRead, now)

	for _, transformHandler := range self.transformHandlers {
		transformHandler.Rx(m, self)
	}

	for _, peekHandler := range self.peekHandlers {
		peekHandler.Rx(m, self)
	}

	handled := false
	if m.IsReply() {
		self.replyCounter++
		if self.replyCounter%100 == 0 && self.waiters.Size() > 1000 {
			self.waiters.reapExpired(now)
		}
		replyFor := m.ReplyFor()
		if replyReceiver := self.waiters.RemoveWaiter(replyFor); replyReceiver != nil {
			log.Tracef("waiter found for message. type [%v], sequence [%v], replyFor [%v]", m.ContentType, m.sequence, replyFor)
			replyReceiver.AcceptReply(m)
			handled = true
		} else {
			log.Debugf("no waiter for message. type [%v], sequence [%v], replyFor [%v]", m.ContentType, m.sequence, replyFor)
		}
	}

	if !handled {
		if receiveHandler, found := self.receiveHandlers[m.ContentType]; found {
			receiveHandler.HandleReceive(m, self)

		} else if anyHandler, found := self.receiveHandlers[AnyContentType]; found {
			anyHandler.HandleReceive(m, self)
		} else {
			log.Warnf("dropped message. type [%d], sequence [%v], replyFor [%v]", m.ContentType, m.sequence, m.ReplyFor())
		}
	}
}

func (self *multiChannelImpl) Tx(underlay Underlay, sendable Sendable, writeTimeout time.Duration) error {
	log := pfxlog.ContextLogger(self.Label())

	sendListener := sendable.SendListener()
	m := sendable.Msg()

	if err := sendable.Context().Err(); err != nil {
		sendListener.NotifyErr(TimeoutError{err})
		return nil
	}

	sendListener.NotifyBeforeWrite()

	if m == nil { // allow nil message in Sendable so we can send tracers to check time from send to write
		return nil
	}

	for _, transformHandler := range self.transformHandlers {
		transformHandler.Tx(m, self)
	}

	self.waiters.AddWaiter(sendable)

	var err error
	if writeTimeout > 0 {
		if err = underlay.SetWriteTimeout(writeTimeout); err != nil {
			log.WithError(err).Errorf("unable to set write timeout")
			sendListener.NotifyErr(err)
			return err
		}
	}

	err = underlay.Tx(m)

	if err != nil {
		log.WithError(err).Errorf("write error")
		self.waiters.RemoveWaiter(m.sequence)

		for _, errorHandler := range self.errorHandlers {
			errorHandler.HandleError(err, self)
		}

		// if we were able to requeue it, don't cancel sendable
		if !self.underlayHandler.HandleTxFailed(underlay, sendable) {
			sendListener.NotifyErr(err)
			sendListener.NotifyAfterWrite()
		}

		return err
	}

	for _, peekHandler := range self.peekHandlers {
		peekHandler.Tx(m, self)
	}

	sendListener.NotifyAfterWrite()

	return nil
}

func (self *multiChannelImpl) closeUnderlay(underlay Underlay, notifier *CloseNotifier) {
	self.lock.Lock()
	if err := underlay.Close(); err != nil {
		pfxlog.Logger().WithField("context", self.Label()).WithError(err).Error("error closing underlay")
	}

	notifier.NotifyClosed()

	underlayRemoved := false
	self.underlays.DeleteIf(func(element Underlay) bool {
		if underlay == element {
			underlayRemoved = true
			return true
		}
		return false
	})
	if *self.fallbackUnderlay.Load() == underlay {
		underlays := self.underlays.Value()
		if len(underlays) > 0 {
			lastUnderlay := underlays[len(underlays)-1]
			self.fallbackUnderlay.Store(&lastUnderlay)
		}
	}
	self.lock.Unlock()

	if underlayRemoved {
		self.underlayHandler.HandleUnderlayClose(self, underlay)
	}
}

func (self *multiChannelImpl) GetTimeSinceLastRead() time.Duration {
	return time.Duration(info.NowInMilliseconds()-atomic.LoadInt64(&self.lastRead)) * time.Millisecond
}

func (self *multiChannelImpl) Txer(underlay Underlay, notifier *CloseNotifier) {
	defer self.closeUnderlay(underlay, notifier)

	log := pfxlog.ContextLogger(self.Label())

	var writeTimeout time.Duration
	if options := self.GetOptions(); options != nil {
		writeTimeout = options.WriteTimeout
	}

	messageSource := self.underlayHandler.GetMessageSource(underlay)

	for {
		sendable, err := messageSource(notifier)
		if err != nil {
			return
		}

		if err = self.Tx(underlay, sendable, writeTimeout); err != nil {
			if self.IsClosed() {
				log.WithError(err).Debug("tx error")
			} else {
				log.WithError(err).Error("tx error")
			}
			return
		}
	}
}

func (self *multiChannelImpl) Rxer(underlay Underlay, notifier *CloseNotifier) {
	defer self.closeUnderlay(underlay, notifier)

	log := pfxlog.ContextLogger(self.Label())
	log.Debug("started")
	defer log.Debug("exited")

	underlayType := GetUnderlayType(underlay)
	injectType := self.flags.IsSet(flagInjectUnderlayType)

	for {
		m, err := underlay.Rx()
		if err != nil {
			if err == io.EOF {
				log.WithError(err).Debug("EOF")
			} else if self.IsClosed() {
				log.WithError(err).Debug("rx error")
			} else {
				log.WithError(err).Error("rx error")
			}
			return
		}

		if injectType {
			m.Headers.PutStringHeader(UnderlayTypeHeader, underlayType)
		}
		self.Rx(m)
	}
}

func (self *multiChannelImpl) DialUnderlay(factory GroupedUnderlayFactory, underlayType string) {
	log := pfxlog.ContextLogger(self.Label()).WithField("underlayType", underlayType)
	attempt := 1
	for {
		if self.IsClosed() {
			log.Info("multi-underlay channel closed, abandoning dial")
			return
		}

		dialTimeout := self.GetOptions().ConnectTimeout
		if dialTimeout == 0 {
			dialTimeout = DefaultConnectTimeout
		}

		underlay, err := factory.CreateGroupedUnderlay(self.ConnectionId(), self.groupSecret, underlayType, dialTimeout)
		if err == nil {
			if err = self.AcceptUnderlay(underlay); err != nil {
				log.WithError(err).Error("dial of new underlay failed")
				factory.DialFailed(self, underlayType, attempt)
			}
			return
		} else {
			factory.DialFailed(self, underlayType, attempt)
		}
		attempt++
	}
}

func GetUnderlayType(underlay Underlay) string {
	return string(underlay.Headers()[TypeHeader])
}

type underlayConstraint struct {
	numDesired int
	minAllowed int
}

type UnderlayConstraints struct {
	types           map[string]underlayConstraint
	minTotal        uint32
	applyInProgress atomic.Bool
	lastDial        concurrenz.AtomicValue[time.Time]
}

func (self *UnderlayConstraints) LastDialTime() time.Time {
	return self.lastDial.Load()
}

func (self *UnderlayConstraints) SetMinTotal(minTotal uint32) {
	self.minTotal = minTotal
}

func (self *UnderlayConstraints) AddConstraint(underlayType string, numDesired int, minAllowed int) {
	if self.types == nil {
		self.types = make(map[string]underlayConstraint)
	}
	self.types[underlayType] = underlayConstraint{numDesired, minAllowed}
}

func (self *UnderlayConstraints) CheckStateValid(ch MultiChannel, close bool) bool {
	counts := ch.GetUnderlayCountsByType()
	return self.countsShowValidState(ch, counts, close)
}

func (self *UnderlayConstraints) countsShowValidState(ch MultiChannel, counts map[string]int, close bool) bool {
	for underlayType, constraint := range self.types {
		if constraint.minAllowed > counts[underlayType] {
			if close {
				pfxlog.Logger().WithField("conn", ch.LogicalName()).
					WithField("channelId", ch.ConnectionId()).
					WithField("label", ch.Label()).
					WithField("underlays", counts).
					Infof("not enough open underlays of type '%s', closing multi-underlay channel", underlayType)
				if err := ch.Close(); err != nil {
					pfxlog.Logger().WithError(err).Error("error closing underlay")
				}
			}
			return false
		}
	}

	totalCount := 0
	for _, count := range counts {
		totalCount += count
	}

	if uint32(totalCount) < self.minTotal {
		if close {
			pfxlog.Logger().WithField("conn", ch.LogicalName()).
				WithField("channelId", ch.ConnectionId()).
				WithField("label", ch.Label()).
				WithField("underlays", counts).
				Info("not enough total open underlays, closing multi-underlay channel")
			if err := ch.Close(); err != nil {
				pfxlog.Logger().WithError(err).Error("error closing channel")
			}
		}
		return false
	}

	return true
}

func (self *UnderlayConstraints) Apply(ch MultiChannel, factory GroupedUnderlayFactory) {
	log := pfxlog.Logger().WithField("conn", ch.Label())
	log.Debug("starting constraint check")

	if ch.IsClosed() {
		return
	}

	if !self.CheckStateValid(ch, true) {
		return
	}

	if !self.applyInProgress.CompareAndSwap(false, true) {
		return
	}

	defer self.applyInProgress.Store(false)

	for !ch.IsClosed() {
		counts := ch.GetUnderlayCountsByType()
		if !self.countsShowValidState(ch, counts, true) {
			return
		}

		allSatisfied := true
		for underlayType, constraint := range self.types {
			log.WithField("underlayType", underlayType).
				WithField("numDesired", constraint.numDesired).
				WithField("current", counts[underlayType]).
				Debug("checking constraint")
			if constraint.numDesired > counts[underlayType] {
				dialElapsed := time.Since(self.lastDial.Load())

				log.WithField("conn", ch.Label()).
					WithField("underlayType", underlayType).
					WithField("timeSinceLastDial", dialElapsed.String()).
					Info("additional connections desired, dialing...")

				allSatisfied = false

				ch.DialUnderlay(factory, underlayType)
				self.lastDial.Store(time.Now())
			}
		}

		if allSatisfied {
			pfxlog.Logger().WithField("conn", ch.Label()).Debug("constraints satisfied")
			return
		}
	}
}

func NewCloseNotifier() *CloseNotifier {
	return &CloseNotifier{
		c: make(chan struct{}),
	}
}

type CloseNotifier struct {
	c        chan struct{}
	notified atomic.Bool
}

func (self *CloseNotifier) NotifyClosed() {
	if self.notified.CompareAndSwap(false, true) {
		close(self.c)
	}
}

func (self *CloseNotifier) GetCloseNotify() <-chan struct{} {
	return self.c
}
