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

package edge

import (
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"math"
	"sync"
	"sync/atomic"
)

type MsgSink interface {
	HandleMuxClose() error
	Id() uint32
	Accept(msg *channel.Message)
}

type MsgMux interface {
	channel.TypedReceiveHandler
	channel.CloseHandler
	AddMsgSink(sink MsgSink) error
	RemoveMsgSink(sink MsgSink)
	RemoveMsgSinkById(sinkId uint32)
	Close()
	GetNextId() uint32
}

func NewCowMapMsgMux() MsgMux {
	result := &CowMapMsgMux{
		maxId: (math.MaxUint32 / 2) - 1,
	}
	result.sinks.Store(map[uint32]MsgSink{})
	return result
}

type CowMapMsgMux struct {
	sync.Mutex
	closed atomic.Bool
	sinks  atomic.Value
	nextId uint32
	minId  uint32
	maxId  uint32
}

func (mux *CowMapMsgMux) GetNextId() uint32 {
	nextId := atomic.AddUint32(&mux.nextId, 1)
	sinks := mux.getSinks()
	for {
		if _, found := sinks[nextId]; found {
			// if it's in use, try next one
			nextId = atomic.AddUint32(&mux.nextId, 1)
		} else if nextId < mux.minId || nextId >= mux.maxId {
			// it's not in use, but not in the valid range, so reset to beginning of range
			atomic.StoreUint32(&mux.nextId, mux.minId)
			nextId = atomic.AddUint32(&mux.nextId, 1)
		} else {
			// If it's not in use, and in the valid range, return it
			return nextId
		}
	}
}

func (mux *CowMapMsgMux) ContentType() int32 {
	return ContentTypeData
}

func (mux *CowMapMsgMux) HandleReceive(msg *channel.Message, ch channel.Channel) {
	connId, found := msg.GetUint32Header(ConnIdHeader)
	if !found {
		pfxlog.Logger().Errorf("received edge message with no connId header. content type: %v", msg.ContentType)
		return
	}

	sinks := mux.getSinks()
	if sink, found := sinks[connId]; found {
		sink.Accept(msg)
	} else if msg.ContentType == ContentTypeConnInspectRequest {
		pfxlog.Logger().WithField("connId", connId).Trace("no conn found for connection inspect")
		resp := NewConnInspectResponse(connId, ConnTypeInvalid, fmt.Sprintf("invalid conn id [%v]", connId))
		if err := resp.ReplyTo(msg).Send(ch); err != nil {
			logrus.WithFields(GetLoggerFields(msg)).WithError(err).
				Error("failed to send inspect response")
		}
	} else {
		pfxlog.Logger().Debugf("unable to dispatch msg received for unknown edge conn id: %v", connId)
	}
}

func (mux *CowMapMsgMux) HandleClose(channel.Channel) {
	mux.Close()
}

func (mux *CowMapMsgMux) AddMsgSink(sink MsgSink) error {
	if mux.closed.Load() {
		return errors.Errorf("mux is closed, can't add sink with id [%v]", sink.Id())
	}

	var err error
	mux.updateSinkMap(func(m map[uint32]MsgSink) {
		if _, found := m[sink.Id()]; found {
			err = errors.Errorf("sink id %v already in use", sink.Id())
		} else {
			m[sink.Id()] = sink
		}
	})

	// check again, just in case it was closed while we were adding
	if mux.closed.Load() {
		return errors.Errorf("mux is closed, can't add sink with id [%v]", sink.Id())
	}

	return err
}

func (mux *CowMapMsgMux) RemoveMsgSink(sink MsgSink) {
	mux.RemoveMsgSinkById(sink.Id())
}

func (mux *CowMapMsgMux) RemoveMsgSinkById(sinkId uint32) {
	mux.updateSinkMap(func(m map[uint32]MsgSink) {
		delete(m, sinkId)
	})
}

func (mux *CowMapMsgMux) updateSinkMap(f func(map[uint32]MsgSink)) {
	mux.Lock()
	defer mux.Unlock()

	current := mux.getSinks()
	result := map[uint32]MsgSink{}
	for k, v := range current {
		result[k] = v
	}
	f(result)
	mux.sinks.Store(result)
}

func (mux *CowMapMsgMux) Close() {
	if mux.closed.CompareAndSwap(false, true) {
		// we don't need to lock the mux because due to the atomic bool, only one go-routine will enter this.
		// If the sink HandleMuxClose methods do anything with the mux, like remove themselves, they will acquire
		// their own locks
		sinks := mux.getSinks()
		for _, val := range sinks {
			if err := val.HandleMuxClose(); err != nil {
				pfxlog.Logger().
					WithField("sinkId", val.Id()).
					WithError(err).
					Error("error while closing message sink")
			}
		}
	}
}

func (mux *CowMapMsgMux) getSinks() map[uint32]MsgSink {
	return mux.sinks.Load().(map[uint32]MsgSink)
}
