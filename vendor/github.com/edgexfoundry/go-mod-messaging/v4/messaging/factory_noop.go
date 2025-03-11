//go:build no_messagebus
// +build no_messagebus

//
// Copyright (c) 2021 Intel Corporation
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

package messaging

import (
	"errors"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// NewMessageClient is noop implementation when service doesn't need the message bus.
// This is need when this module is included in the common go-mod-bootstrap, but some service
// such as security service have no need for messaging.
func NewMessageClient(msgConfig types.MessageBusConfig) (MessageClient, error) {
	return nil, errors.New("messaging was disabled during build with the no_messagebus build flag")
}
