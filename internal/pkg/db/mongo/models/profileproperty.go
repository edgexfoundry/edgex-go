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

type ProfileProperty struct {
	Value PropertyValue `bson:"value"`
	Units Units         `bson:"units"`
}

func (p *ProfileProperty) ToContract() (c contract.ProfileProperty) {
	c.Value = p.Value.ToContract()
	c.Units = p.Units.ToContract()

	return
}

func (p *ProfileProperty) FromContract(c contract.ProfileProperty) {
	p.Value.FromContract(c.Value)
	p.Units.FromContract(c.Units)
}
