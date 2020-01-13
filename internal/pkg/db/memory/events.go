/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package memory

import (
	correlation "github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func (c *Client) Events() ([]contract.Event, error) {
	return []contract.Event{}, nil
}

func (c *Client) EventsWithLimit(limit int) ([]contract.Event, error) {
	return []contract.Event{}, nil
}

func (c *Client) AddEvent(e correlation.Event) (string, error) {
	return "", nil
}

func (c *Client) UpdateEvent(e correlation.Event) error {
	return nil
}

func (c *Client) EventById(id string) (contract.Event, error) {
	return contract.Event{}, nil
}

func (c *Client) EventsByChecksum(checksum string) ([]contract.Event, error) {
	return []contract.Event{}, nil
}

func (c *Client) EventCount() (int, error) {
	return 0, nil
}

func (c *Client) EventCountByDeviceId(id string) (int, error) {
	return 0, nil
}

func (c *Client) DeleteEventById(id string) error {
	return nil
}

func (c *Client) DeleteEventsByDevice(deviceId string) (int, error) {
	return 0, nil
}

func (c *Client) EventsForDeviceLimit(id string, limit int) ([]contract.Event, error) {
	return []contract.Event{}, nil
}

func (c *Client) EventsForDevice(id string) ([]contract.Event, error) {
	return []contract.Event{}, nil
}

func (c *Client) EventsByCreationTime(startTime, endTime int64, limit int) ([]contract.Event, error) {
	return []contract.Event{}, nil
}

func (c *Client) EventsOlderThanAge(age int64) ([]contract.Event, error) {
	return []contract.Event{}, nil
}

func (c *Client) EventsPushed() ([]contract.Event, error) {
	return []contract.Event{}, nil
}

func (c *Client) ScrubAllEvents() error {
	return nil
}
