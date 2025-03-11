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
	"github.com/edgexfoundry/go-mod-messaging/v4/internal/pkg/nats/interfaces"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
	"github.com/nats-io/nats.go"
)

const (
	DeliverAll            = "all"
	DeliverLast           = "last"
	DeliverLastPerSubject = "lastpersubject"
	DeliverNew            = "new"
)

func newConnection(cc natsMessaging.ClientConfig) (interfaces.Connection, error) {
	co, err := cc.ConnectOpt()

	if err != nil {
		return nil, err
	}

	conn, err := nats.Connect(cc.BrokerURL, co...)

	if err != nil {
		return nil, err
	}

	js, err := conn.JetStream()

	if err != nil {
		return nil, err
	}

	if cc.AutoProvision {
		if apErr := autoProvision(cc, js); apErr != nil {
			return nil, apErr
		}
	}

	return &connection{cc, conn, js, subOpt(cc), pubOpt(cc)}, nil
}

func autoProvision(cc natsMessaging.ClientConfig, js nats.JetStreamContext) error {
	streamName := cc.Durable

	autoProvisionSubject := natsMessaging.TopicToSubject(cc.Subject)

	if strings.TrimSpace(streamName) == "" {
		// fall back to formatted subject if no durable specified
		streamName = subjectToStreamName(autoProvisionSubject)
	}

	// only need to check for existence here
	_, err := js.StreamInfo(streamName)

	if err != nil {
		// only interested if an error encountered on stream provisioning
		_, err = js.AddStream(&nats.StreamConfig{
			Name:        streamName,
			Description: "",
			Subjects:    []string{autoProvisionSubject},
		})
	}

	return err
}

// NewClient creates a new client using NATS JetStream.
func NewClient(cfg types.MessageBusConfig) (*natsMessaging.Client, error) {
	return natsMessaging.NewClientWithConnectionFactory(cfg, newConnection)
}

func subOpt(cc natsMessaging.ClientConfig) []nats.SubOpt {
	return []nats.SubOpt{nats.AckExplicit(), parseDeliver(cc.Deliver)()}
}

func pubOpt(cc natsMessaging.ClientConfig) []nats.PubOpt {
	return []nats.PubOpt{nats.RetryAttempts(cc.DefaultPubRetryAttempts)}
}

// parseDeliver will return the appropriate nats delivery option function for a given string.
// it returns the func itself NOT the result of invocation so that results can be asserted
// independently - invocation to get the underlying option is left as an exercise for the caller.
func parseDeliver(configured string) func() nats.SubOpt {
	switch strings.ToLower(configured) {
	case DeliverAll:
		return nats.DeliverAll
	case DeliverLast:
		return nats.DeliverLast
	case DeliverLastPerSubject:
		return nats.DeliverLastPerSubject
	default:
		// DeliverNew is default for parity with core NATS
		return nats.DeliverNew
	}
}
