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

//go:build !include_nats_messaging

package jetstream

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-messaging/v4/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// NewClient initializes creates a new client using NATS core
func NewClient(_ types.MessageBusConfig) (*pkg.NoopClient, error) {
	return nil, fmt.Errorf("to enable NATS message bus options please build using the flag include_nats_messaging")
}
