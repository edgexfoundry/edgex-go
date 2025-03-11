/*
	Copyright 2019 NetFoundry Inc.

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

package network

import (
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/pkg/errors"
	"math"
	"net"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type baseListener struct {
	service *rest_model.ServiceDetail
	acceptC chan edge.Conn
	errorC  chan error
	closed  atomic.Bool
}

func (listener *baseListener) Network() string {
	return "ziti"
}

func (listener *baseListener) String() string {
	return *listener.service.Name
}

func (listener *baseListener) Addr() net.Addr {
	return listener
}

func (listener *baseListener) IsClosed() bool {
	return listener.closed.Load()
}

func (listener *baseListener) Accept() (net.Conn, error) {
	conn, err := listener.AcceptEdge()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (listener *baseListener) AcceptEdge() (edge.Conn, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for !listener.closed.Load() {
		select {
		case conn, ok := <-listener.acceptC:
			if ok && conn != nil {
				return conn, nil
			} else {
				listener.closed.Store(true)
			}
		case <-ticker.C:
		}
	}

	select {
	case err := <-listener.errorC:
		return nil, fmt.Errorf("listener is closed (%w)", err)
	default:
	}

	return nil, errors.New("listener is closed")
}

type edgeListener struct {
	baseListener
	token       string
	edgeChan    *edgeConn
	manualStart bool
	established atomic.Bool
	eventC      chan *edge.ListenerEvent
}

func (listener *edgeListener) Id() uint32 {
	return listener.edgeChan.Id()
}

func (listener *edgeListener) UpdateCost(cost uint16) error {
	return listener.updateCostAndPrecedence(&cost, nil)
}

func (listener *edgeListener) UpdatePrecedence(precedence edge.Precedence) error {
	return listener.updateCostAndPrecedence(nil, &precedence)
}

func (listener *edgeListener) UpdateCostAndPrecedence(cost uint16, precedence edge.Precedence) error {
	return listener.updateCostAndPrecedence(&cost, &precedence)
}

func (listener *edgeListener) updateCostAndPrecedence(cost *uint16, precedence *edge.Precedence) error {
	logger := pfxlog.Logger().
		WithField("connId", listener.edgeChan.Id()).
		WithField("serviceName", listener.edgeChan.serviceName).
		WithField("session", listener.token)

	logger.Debug("sending update bind request to edge router")
	request := edge.NewUpdateBindMsg(listener.edgeChan.Id(), listener.token, cost, precedence)
	listener.edgeChan.TraceMsg("updateCostAndPrecedence", request)
	return request.WithTimeout(5 * time.Second).SendAndWaitForWire(listener.edgeChan.Channel)
}

func (listener *edgeListener) SendHealthEvent(pass bool) error {
	logger := pfxlog.Logger().
		WithField("connId", listener.edgeChan.Id()).
		WithField("serviceName", listener.edgeChan.serviceName).
		WithField("session", listener.token).
		WithField("health.status", pass)

	logger.Debug("sending health event to edge router")
	request := edge.NewHealthEventMsg(listener.edgeChan.Id(), listener.token, pass)
	listener.edgeChan.TraceMsg("healthEvent", request)
	return request.WithTimeout(5 * time.Second).SendAndWaitForWire(listener.edgeChan.Channel)
}

func (listener *edgeListener) Close() error {
	return listener.close(true)
}

func (listener *edgeListener) close(closedByRemote bool) error {
	if !listener.closed.CompareAndSwap(false, true) {
		// already closed
		return nil
	}

	edgeChan := listener.edgeChan

	logger := pfxlog.Logger().
		WithField("connId", listener.edgeChan.Id()).
		WithField("sessionId", listener.token)

	logger.Debug("removing listener for session")
	edgeChan.hosting.Remove(listener.token)

	defer func() {
		edgeChan.close(closedByRemote)
		listener.acceptC <- nil // signal listeners that listener is closed
	}()

	unbindRequest := edge.NewUnbindMsg(edgeChan.Id(), listener.token)
	listener.edgeChan.TraceMsg("close", unbindRequest)
	if err := unbindRequest.WithTimeout(5 * time.Second).SendAndWaitForWire(edgeChan.Channel); err != nil {
		logger.WithError(err).Error("unable to unbind session for conn")
		return err
	}

	return nil
}

type MultiListener interface {
	edge.Listener
	AddListener(listener edge.Listener, closeHandler func())
	NotifyOfChildError(err error)
	GetServiceName() string
	GetService() *rest_model.ServiceDetail
	CloseWithError(err error)
	GetEstablishedCount() uint
}

func NewMultiListener(service *rest_model.ServiceDetail, getSessionF func() *rest_model.SessionDetail) MultiListener {
	return &multiListener{
		baseListener: baseListener{
			service: service,
			acceptC: make(chan edge.Conn),
			errorC:  make(chan error),
		},
		listeners:      map[*edgeListener]struct{}{},
		getSessionF:    getSessionF,
		listenerEventC: make(chan *edge.ListenerEvent, 3),
	}
}

type multiListener struct {
	baseListener
	listeners            map[*edgeListener]struct{}
	listenerLock         sync.Mutex
	getSessionF          func() *rest_model.SessionDetail
	listenerEventHandler atomic.Value
	errorEventHandler    atomic.Value
	listenerEventC       chan *edge.ListenerEvent
}

func (self *multiListener) Id() uint32 {
	return math.MaxUint32
}

func (self *multiListener) GetEstablishedCount() uint {
	var count uint
	self.listenerLock.Lock()
	defer self.listenerLock.Unlock()
	for v := range self.listeners {
		if v.established.Load() {
			count++
		}
	}
	return count
}

func (self *multiListener) SetConnectionChangeHandler(handler func([]edge.Listener)) {
	self.listenerEventHandler.Store(handler)

	self.listenerLock.Lock()
	defer self.listenerLock.Unlock()
	self.notifyOfConnectionChange()
}

func (self *multiListener) GetConnectionChangeHandler() func([]edge.Listener) {
	val := self.listenerEventHandler.Load()
	if val == nil {
		return nil
	}
	return val.(func([]edge.Listener))
}

func (self *multiListener) SetErrorEventHandler(handler func(error)) {
	self.errorEventHandler.Store(handler)
}

func (self *multiListener) GetErrorEventHandler() func(error) {
	val := self.errorEventHandler.Load()
	if val == nil {
		return nil
	}
	return val.(func(error))
}

func (self *multiListener) NotifyOfChildError(err error) {
	pfxlog.Logger().Infof("notify error handler of error: %v", err)
	if handler := self.GetErrorEventHandler(); handler != nil {
		handler(err)
	}
}

func (self *multiListener) notifyOfConnectionChange() {
	if handler := self.GetConnectionChangeHandler(); handler != nil {
		var list []edge.Listener
		for k := range self.listeners {
			list = append(list, k)
		}
		go handler(list)
	}
}

func (self *multiListener) GetCurrentSession() *rest_model.SessionDetail {
	return self.getSessionF()
}

func (self *multiListener) UpdateCost(cost uint16) error {
	self.listenerLock.Lock()
	defer self.listenerLock.Unlock()

	var resultErrors []error
	for child := range self.listeners {
		if err := child.UpdateCost(cost); err != nil {
			resultErrors = append(resultErrors, err)
		}
	}
	return self.condenseErrors(resultErrors)
}

func (self *multiListener) UpdatePrecedence(precedence edge.Precedence) error {
	self.listenerLock.Lock()
	defer self.listenerLock.Unlock()

	var resultErrors []error
	for child := range self.listeners {
		if err := child.UpdatePrecedence(precedence); err != nil {
			resultErrors = append(resultErrors, err)
		}
	}
	return self.condenseErrors(resultErrors)
}

func (self *multiListener) UpdateCostAndPrecedence(cost uint16, precedence edge.Precedence) error {
	self.listenerLock.Lock()
	defer self.listenerLock.Unlock()

	var resultErrors []error
	for child := range self.listeners {
		if err := child.UpdateCostAndPrecedence(cost, precedence); err != nil {
			resultErrors = append(resultErrors, err)
		}
	}
	return self.condenseErrors(resultErrors)
}

func (self *multiListener) SendHealthEvent(pass bool) error {
	self.listenerLock.Lock()
	defer self.listenerLock.Unlock()

	// only send to first child, otherwise we get duplicate event reporting
	for child := range self.listeners {
		return child.SendHealthEvent(pass)
	}
	return nil
}

func (self *multiListener) condenseErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	if len(errors) == 1 {
		return errors[0]
	}
	return MultipleErrors(errors)
}

func (self *multiListener) GetServiceName() string {
	return *self.service.Name
}

func (self *multiListener) GetService() *rest_model.ServiceDetail {
	return self.service
}

func (self *multiListener) AddListener(netListener edge.Listener, closeHandler func()) {
	if self.closed.Load() {
		return
	}

	edgeListener, ok := netListener.(*edgeListener)
	if !ok {
		pfxlog.Logger().Errorf("multi-listener expects only listeners created by the SDK, not %v", reflect.TypeOf(self))
		return
	}

	self.listenerLock.Lock()
	defer self.listenerLock.Unlock()
	self.listeners[edgeListener] = struct{}{}

	closer := func() {
		self.listenerLock.Lock()
		defer self.listenerLock.Unlock()
		delete(self.listeners, edgeListener)

		self.notifyOfConnectionChange()
		go closeHandler()
	}

	self.notifyOfConnectionChange()

	go self.forward(edgeListener, closer)
}

func (self *multiListener) forward(edgeListener *edgeListener, closeHandler func()) {
	defer func() {
		if err := edgeListener.Close(); err != nil {
			pfxlog.Logger().Errorf("failure closing edge listener: (%v)", err)
		}
		closeHandler()
	}()

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for !self.closed.Load() && !edgeListener.closed.Load() {
		select {
		case conn, ok := <-edgeListener.acceptC:
			if !ok || conn == nil {
				// closed, returning
				return
			}
			self.accept(conn, ticker)
		case <-ticker.C:
			// lets us check if the listener is closed, and exit if it has
		}
	}
}

func (self *multiListener) accept(conn edge.Conn, ticker *time.Ticker) {
	for !self.closed.Load() {
		select {
		case self.acceptC <- conn:
			return
		case <-ticker.C:
			// lets us check if the listener is closed, and exit if it has
		}
	}
}

func (self *multiListener) Close() error {
	self.closed.Store(true)

	self.listenerLock.Lock()
	defer self.listenerLock.Unlock()

	var resultErrors []error
	for child := range self.listeners {
		if err := child.Close(); err != nil {
			resultErrors = append(resultErrors, err)
		}
	}

	self.listeners = nil

	select {
	case self.acceptC <- nil:
	default:
		// If the queue is full, bail out, we're just popping a nil on the
		// accept queue to let it return from accept more quickly
	}

	return self.condenseErrors(resultErrors)
}

func (self *multiListener) CloseWithError(err error) {
	select {
	case self.errorC <- err:
	default:
	}

	self.closed.Store(true)
}

type MultipleErrors []error

func (e MultipleErrors) Error() string {
	if len(e) == 0 {
		return "no errors occurred"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	buf := strings.Builder{}
	buf.WriteString("multiple errors occurred")
	for idx, err := range e {
		buf.WriteString(fmt.Sprintf(" %v: %v", idx, err))
	}
	return buf.String()
}
