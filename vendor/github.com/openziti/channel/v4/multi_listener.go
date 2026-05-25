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
	"errors"
	"sync"

	"github.com/michaelquigley/pfxlog"
)

type MultiChannelFactory func(underlay Underlay, closeCallback func()) (MultiChannel, error)
type UngroupedChannelFallback func(underlay Underlay) error

type MultiListener struct {
	channels                 map[string]MultiChannel
	lock                     sync.Mutex
	multiChannelFactory      MultiChannelFactory
	ungroupedChannelFallback UngroupedChannelFallback
}

func (self *MultiListener) AcceptUnderlay(underlay Underlay) {
	isGrouped, _ := Headers(underlay.Headers()).GetBoolHeader(IsGroupedHeader)

	log := pfxlog.Logger().
		WithField("underlayId", underlay.ConnectionId()).
		WithField("underlayType", GetUnderlayType(underlay)).
		WithField("isGrouped", isGrouped)

	if !isGrouped {
		if err := self.ungroupedChannelFallback(underlay); err != nil {
			log.WithError(err).Error("failed to create channel")
			if closeErr := underlay.Close(); closeErr != nil {
				log.WithError(closeErr).Error("error closing underlay")
			}
		}
		return
	}

	chId := underlay.ConnectionId()
	isFirst, _ := Headers(underlay.Headers()).GetBoolHeader(IsFirstGroupConnection)

	self.lock.Lock()
	mc, ok := self.channels[chId]
	self.lock.Unlock()

	if ok {
		log.Info("found existing channel for underlay")
		if err := mc.AcceptUnderlay(underlay); err != nil {
			log.WithError(err).Error("error accepting underlay")
		}
	} else {
		if !isFirst {
			log.Info("no existing channel found for underlay, but isFirstGroupConnection not set, closing connection")
			if err := underlay.Close(); err != nil {
				log.WithError(err).Error("error closing underlay")
			}
			return
		}

		log.Info("no existing channel found for underlay")
		var err error
		mc, err = self.multiChannelFactory(underlay, func() {
			self.CloseChannel(chId)
		})

		if mc == nil && err == nil {
			err = errors.New("multi-channel factory returned nil")
		}

		if err != nil {
			log.WithError(err).Error("failed to create multi-underlay channel")
			if closeErr := underlay.Close(); closeErr != nil {
				log.WithError(closeErr).Error("error closing underlay")
			}
		} else {
			self.lock.Lock()
			self.channels[chId] = mc
			self.lock.Unlock()
		}
	}
}

func (self *MultiListener) CloseChannel(chId string) {
	self.lock.Lock()
	delete(self.channels, chId)
	self.lock.Unlock()
}

func NewMultiListener(channelF MultiChannelFactory, fallback UngroupedChannelFallback) *MultiListener {
	result := &MultiListener{
		channels:                 make(map[string]MultiChannel),
		multiChannelFactory:      channelF,
		ungroupedChannelFallback: fallback,
	}
	return result
}
