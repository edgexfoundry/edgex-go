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

func (c *Client) ValueDescriptors() ([]contract.ValueDescriptor, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) AddValueDescriptor(v contract.ValueDescriptor) (string, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) UpdateValueDescriptor(cvd contract.ValueDescriptor) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteValueDescriptorById(id string) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ValueDescriptorByName(name string) (contract.ValueDescriptor, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ValueDescriptorsByName(names []string) ([]contract.ValueDescriptor, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ValueDescriptorById(id string) (contract.ValueDescriptor, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ValueDescriptorsByUomLabel(uomLabel string) ([]contract.ValueDescriptor, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ValueDescriptorsByLabel(label string) ([]contract.ValueDescriptor, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ValueDescriptorsByType(t string) ([]contract.ValueDescriptor, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ScrubAllValueDescriptors() error {
	panic(UnimplementedMethodPanicMessage)
}
