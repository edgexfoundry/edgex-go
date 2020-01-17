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
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) EventsWithLimit(limit int) ([]contract.Event, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) AddEvent(e correlation.Event) (string, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) UpdateEvent(e correlation.Event) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) EventById(id string) (contract.Event, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) EventsByChecksum(checksum string) ([]contract.Event, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) EventCount() (int, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) EventCountByDeviceId(id string) (int, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteEventById(id string) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteEventsByDevice(deviceId string) (int, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) EventsForDeviceLimit(id string, limit int) ([]contract.Event, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) EventsForDevice(id string) ([]contract.Event, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) EventsByCreationTime(startTime, endTime int64, limit int) ([]contract.Event, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) EventsOlderThanAge(age int64) ([]contract.Event, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) EventsPushed() ([]contract.Event, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ScrubAllEvents() error {
	panic(UnimplementedMethodPanicMessage)
}
