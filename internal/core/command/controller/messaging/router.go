//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"errors"
	"sync"
)

// MessagingRouter defines interface for Command Service to know
// where to route the receiving device command response.
type MessagingRouter interface {
	// ResponseTopic returns the responseTopicPrefix by requestId, and a boolean value
	// indicates its original source(external MQTT or internal MessageBus).
	ResponseTopic(requestId string) (string, bool, error)
	// SetResponseTopic sets the responseTopicPrefix with RequestId as the key
	SetResponseTopic(requestId string, topic string, external bool)
}

func NewMessagingRouter() MessagingRouter {
	return &router{
		internalCommandRequestMap: make(map[string]string),
		externalCommandRequestMap: make(map[string]string),
	}
}

type router struct {
	mutex                     sync.Mutex
	internalCommandRequestMap map[string]string
	externalCommandRequestMap map[string]string
}

func (r *router) ResponseTopic(requestId string) (string, bool, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	topic, ok := r.externalCommandRequestMap[requestId]
	if ok {
		delete(r.externalCommandRequestMap, requestId)
		return topic, true, nil
	}

	topic, ok = r.internalCommandRequestMap[requestId]
	if ok {
		delete(r.internalCommandRequestMap, requestId)
		return topic, false, nil
	}

	return "", false, errors.New("requestId not found")
}

func (r *router) SetResponseTopic(requestId string, topic string, external bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if external {
		r.externalCommandRequestMap[requestId] = topic
		return
	}

	r.internalCommandRequestMap[requestId] = topic
}
