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

package jetstream

import (
	"strings"

	natsMessaging "github.com/edgexfoundry/go-mod-messaging/v4/internal/pkg/nats"
	"github.com/nats-io/nats.go"
)

// connection mimics the core NATS publish/subscribe API
// so that NATS and jetstream can use the same client orchestration.
type connection struct {
	cfg     natsMessaging.ClientConfig
	conn    *nats.Conn
	js      nats.JetStreamContext
	subOpts []nats.SubOpt
	pubOpts []nats.PubOpt
}

// Subscribe subscribes to a JetStream subject
func (j connection) QueueSubscribe(s string, q string, handler nats.MsgHandler) (*nats.Subscription, error) {
	opts := j.subOpts
	if strings.TrimSpace(j.cfg.Durable) != "" {
		// use the configured durable name to bind subscription to stream
		opts = append(opts, nats.Durable(j.cfg.Durable))
	} else if strings.TrimSpace(j.cfg.Subject) != "" {
		// use the configured subject to bind subscription to stream
		opts = append(opts, nats.BindStream(subjectToStreamName(natsMessaging.TopicToSubject(j.cfg.Subject))))
	}
	return j.js.QueueSubscribe(s, q, handler, opts...)
}

// PublishMsg publishes a message to JetStream
func (j connection) PublishMsg(msg *nats.Msg) (err error) {
	_, err = j.js.PublishMsg(msg, j.pubOpts...)

	return
}

// Drain will remove all subscription interest and attempt to wait until all messages have finished processing to close and return.
func (j connection) Drain() error {
	return j.conn.Drain()
}
