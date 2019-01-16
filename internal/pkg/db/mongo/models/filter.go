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
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)


type Filter struct {
	DeviceIDs          []string `bson:"deviceIdentifiers,omitempty"`
	ValueDescriptorIDs []string `bson:"valueDescriptorIdentifiers,omitempty"`
}

func (f *Filter) ToContract() (c contract.Filter) {
	c.DeviceIDs = f.DeviceIDs
	c.ValueDescriptorIDs = f.ValueDescriptorIDs

	return
}

func (f *Filter) FromContract(c contract.Filter) {
	f.DeviceIDs = c.DeviceIDs
	f.ValueDescriptorIDs = c.ValueDescriptorIDs
}