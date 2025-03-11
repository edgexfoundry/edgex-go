//
// Copyright (c) 2022 One Track Consulting
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

//go:build include_nats_messaging

package interfaces

import "github.com/nats-io/nats.go"

// Connection provides an interface over basic *nats.Conn methods that we need to interact with the broker
type Connection interface {
	// QueueSubscribe subscribes to a NATS subject, equivalent to default Subscribe if queuegroup not supplied.
	QueueSubscribe(string, string, nats.MsgHandler) (*nats.Subscription, error)
	// PublishMsg sends the provided NATS message to the broker.
	PublishMsg(*nats.Msg) error
	// Drain will end all active subscription interest and attempt to wait for in-flight messages to process before closing.
	Drain() error
}
