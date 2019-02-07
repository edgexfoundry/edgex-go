/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package models

import (
	"encoding/json"

	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

type Event struct {
	CorrelationId string `json:"correlation-id"`
	contract.Event
}

// Returns an instance of just the public contract portion of the model Event.
// I don't like returning a pointer from this method but I have to in order to
// satisfy the Filter, Format interfaces.
func (e Event) ToContract() *contract.Event {
	event := contract.Event{
		ID:       e.ID,
		Pushed:   e.Pushed,
		Device:   e.Device,
		Created:  e.Created,
		Modified: e.Modified,
		Origin:   e.Origin,
		Event:    e.Event.Event,
	}

	for _, r := range e.Readings {
		event.Readings = append(event.Readings, r)
	}
	return &event
}

func (e Event) MarshalJSON() ([]byte, error) {
	test := struct {
		CorrelationId *string            `json:"correlation-id,omitempty"`
		ID            *string            `json:"id,omitempty"`
		Pushed        int64              `json:"pushed,omitempty"`
		Device        *string            `json:"device,omitempty"` // Device identifier (name or id)
		Created       int64              `json:"created,omitempty"`
		Modified      int64              `json:"modified,omitempty"`
		Origin        int64              `json:"origin,omitempty"`
		Schedule      *string            `json:"schedule,omitempty"` // Schedule identifier
		Event         *string            `json:"event,omitempty"`    // Schedule event identifier
		Readings      []contract.Reading `json:"readings,omitempty"` // List of readings
	}{
		Pushed:   e.Pushed,
		Created:  e.Created,
		Modified: e.Modified,
		Origin:   e.Origin,
	}

	// Empty strings are null
	if e.CorrelationId != "" {
		test.CorrelationId = &e.CorrelationId
	}

	if e.ID != "" {
		test.ID = &e.ID
	}
	if e.Device != "" {
		test.Device = &e.Device
	}
	if e.Event.Event != "" { //good lord, this is terrible. An Event property on an Event type...
		test.Event = &e.Event.Event
	}

	// Empty arrays are null
	if len(e.Readings) > 0 {
		test.Readings = e.Readings
	}

	return json.Marshal(test)
}
