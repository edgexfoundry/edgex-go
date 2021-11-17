/********************************************************************************
 *  Copyright 2020 Dell Inc.
 *  Copyright (c) 2021 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package redis

import (
	"crypto/tls"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

const (
	// Special identifier used within Redis to signal that a subscriber(consumer) is only interested in the most recent
	// messages after the client has connected. Redis provides other functionality to read all the data from a stream
	// even if has been read previously, which is what we want to avoid for functional consistency with the other
	// implementations of MessageClient.
	LatestStreamMessage = "$"
)

// RedisClientCreator type alias for functions which create RedisClient implementation.
//
// This is mostly used for testing purposes so that we can easily inject mocks.
type RedisClientCreator func(redisServerURL string, password string, tlsConfig *tls.Config) (RedisClient, error)

// RedisClient provides functionality needed to read and send messages to/from Redis' RedisStreams functionality.
//
// The main reason for this interface is to abstract out the underlying client from Client so that it can be mocked and
// allow for easy unit testing. Since 'go-redis' does not leverage interfaces and has complicated entities it can become
// complex to test the operations without requiring a running Redis server.
type RedisClient interface {
	// Send sends a message to the specified topic, aka Publish.
	Send(topic string, message types.MessageEnvelope) error
	// Receive blocking operation which receives the next message for the specified topic, aka Subscribe
	// This supports multi-level topic scheme with wild cards
	Receive(topic string) (*types.MessageEnvelope, error)
	// Close cleans up any entities which need to be deconstructed.
	Close() error
}
