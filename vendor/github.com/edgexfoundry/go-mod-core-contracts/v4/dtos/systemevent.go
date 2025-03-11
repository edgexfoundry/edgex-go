/*******************************************************************************
 * Copyright 2022 Intel Corp.
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

package dtos

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
)

// SystemEvent defines the data for a system event
type SystemEvent struct {
	common.Versionable `json:",inline"`
	Type               string            `json:"type"`
	Action             string            `json:"action"`
	Source             string            `json:"source"`
	Owner              string            `json:"owner"`
	Tags               map[string]string `json:"tags"`
	Details            any               `json:"details"`
	Timestamp          int64             `json:"timestamp"`
}

// NewSystemEvent creates a new SystemEvent for the specified data
func NewSystemEvent(eventType, action, source, owner string, tags map[string]string, details any) SystemEvent {
	return SystemEvent{
		Versionable: common.NewVersionable(),
		Type:        eventType,
		Action:      action,
		Source:      source,
		Owner:       owner,
		Tags:        tags,
		Details:     details,
		Timestamp:   time.Now().UnixNano(),
	}
}

// DecodeDetails decodes the details (any type) into the passed in object
func (s *SystemEvent) DecodeDetails(details any) error {
	if s.Details == nil {
		return errors.New("unable to decode System Event details: Details are nil")
	}

	// Must encode the details to JSON since if the target SystemEvent was decoded from JSON the details are
	// captured in a map[string]interface{}.
	data, err := json.Marshal(s.Details)
	if err != nil {
		return fmt.Errorf("unable to encode System Event details to JSON: %s", err.Error())
	}

	err = json.Unmarshal(data, details)
	if err != nil {
		return fmt.Errorf("unable to decode System Event details from JSON: %s", err.Error())
	}

	return nil
}
