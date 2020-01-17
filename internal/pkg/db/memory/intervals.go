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

import contract "github.com/edgexfoundry/go-mod-core-contracts/models"

func (c *Client) Intervals() ([]contract.Interval, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) IntervalsWithLimit(limit int) ([]contract.Interval, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) IntervalByName(name string) (contract.Interval, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) IntervalById(id string) (contract.Interval, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) AddInterval(interval contract.Interval) (string, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) UpdateInterval(interval contract.Interval) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteIntervalById(id string) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) IntervalActions() ([]contract.IntervalAction, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) IntervalActionsWithLimit(limit int) ([]contract.IntervalAction, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) IntervalActionsByIntervalName(name string) ([]contract.IntervalAction, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) IntervalActionsByTarget(name string) ([]contract.IntervalAction, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) IntervalActionById(id string) (contract.IntervalAction, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) IntervalActionByName(name string) (contract.IntervalAction, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) AddIntervalAction(action contract.IntervalAction) (string, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) UpdateIntervalAction(action contract.IntervalAction) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteIntervalActionById(id string) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ScrubAllIntervalActions() (int, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ScrubAllIntervals() (int, error) {
	panic(UnimplementedMethodPanicMessage)
}
