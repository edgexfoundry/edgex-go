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

import contract "github.com/edgexfoundry/edgex-go/pkg/models"

type DeviceObject struct {
	Description string                 `bson:"description"`
	Name        string                 `bson:"name"`
	Tag         string                 `bson:"tag"`
	Properties  ProfileProperty        `bson:"properties"`
	Attributes  map[string]interface{} `bson:"attributes"`
}

func (do *DeviceObject) ToContract() (c contract.DeviceObject) {
	c.Description = do.Description
	c.Name = do.Name
	c.Tag = do.Tag
	c.Properties = do.Properties.ToContract()
	c.Attributes = do.Attributes

	return
}

func (do *DeviceObject) FromContract(c contract.DeviceObject) {
	do.Description = c.Description
	do.Name = c.Name
	do.Tag = c.Tag
	do.Properties.FromContract(c.Properties)
	do.Attributes = c.Attributes
}
